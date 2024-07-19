package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	ginpprof "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var (
	RestartGameMenuItem                     *fyne.MenuItem
	MusicMenuItem                           *fyne.MenuItem
	AIPlayer1MenuItem                       *fyne.MenuItem
	AIPlayer2MenuItem                       *fyne.MenuItem
	AutoRestartMenuItem                     *fyne.MenuItem
	IncreaseBoardSizeMenuItem               *fyne.MenuItem
	ReduceBoardSizeMenuItem                 *fyne.MenuItem
	ResetBoardSizeMenuItem                  *fyne.MenuItem
	UndoMenuItem                            *fyne.MenuItem
	IncreaseBoardWidthMenuItem              *fyne.MenuItem
	ReduceBoardWidthMenuItem                *fyne.MenuItem
	ResetBoardWidthMenuItem                 *fyne.MenuItem
	IncreaseSearchGoroutinesMenuItem        *fyne.MenuItem
	ReduceSearchGoroutinesMenuItem          *fyne.MenuItem
	ResetSearchGoroutinesMenuItem           *fyne.MenuItem
	ScoreMenuItem                           *fyne.MenuItem
	SaveScreenshotMenuItem                  *fyne.MenuItem
	QuitMenuItem                            *fyne.MenuItem
	HelpMenuItem                            *fyne.MenuItem
	IncreaseAISearchTimeMenuItem            *fyne.MenuItem
	ReduceAISearchTimeMenuItem              *fyne.MenuItem
	ResetAISearchTimeMenuItem               *fyne.MenuItem
	SavePerformanceAnalysisMenuItem         *fyne.MenuItem
	IncreasePerformanceAnalysisTimeMenuItem *fyne.MenuItem
	ReducePerformanceAnalysisTimeMenuItem   *fyne.MenuItem
	ResetPerformanceAnalysisTimeMenuItem    *fyne.MenuItem
)

func GetMessage(head string, value bool) string {
	if value {
		return head + " ON"
	} else {
		return head + " OFF"
	}
}

// getSaveFilePath returns the save file path selected by the user.
func getSaveFilePath() (string, error) {
	var script string

	switch runtime.GOOS {
	case "darwin":
		script = `osascript -e 'set myFile to choose file name with prompt "Save game screenshot as:" default name "dots-and-boxes screenshot.png"' -e 'POSIX path of myFile'`
	case "windows":
		script = `Add-Type -AssemblyName System.Windows.Forms; $file = New-Object System.Windows.Forms.SaveFileDialog; $file.Filter = "PNG Files|*.png"; $file.FileName = "dots-and-boxes screenshot.png"; if($file.ShowDialog() -eq 'OK') {$file.FileName}`
	case "linux":
		script = `zenity --file-selection --save --confirm-overwrite --file-filter="*.png" --filename="dots-and-boxes screenshot.png"`
	default:
		return "", fmt.Errorf("unsupported platform")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	} else {
		cmd = exec.Command("sh", "-c", script)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func init() {
	RestartGameMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(Chess.BoardSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	ScoreMenuItem = &fyne.MenuItem{
		Action: func() {
			message := fmt.Sprintf("Player1 Score: %v\nPlayer2 Score: %v\n", Player1Score, Player2Score)
			Message.Send(message)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyT},
	}

	IncreaseBoardSizeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(Chess.BoardSize + 1)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyEqual},
	}

	ReduceBoardSizeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(Chess.BoardSize - 1)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyMinus},
	}

	ResetBoardSizeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(DefaultBoardSize)
		},
	}

	UndoMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Undo()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyZ},
	}

	SaveScreenshotMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			game.Refresh()
			img := MainWindow.Canvas().Capture()
			buf := new(bytes.Buffer)
			if err := png.Encode(buf, img); err != nil {
				Message.Send(err.Error())
				return
			}
			path, err := getSaveFilePath()
			if err != nil {
				Message.Send(err.Error())
				return
			}
			if err := os.WriteFile(path, buf.Bytes(), 0666); err != nil {
				Message.Send(err.Error())
				return
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS},
	}

	IncreaseBoardWidthMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SetDotDistance(Chess.DotCanvasDistance + 10)
			Message.Send("Now BoardWidth: %v", Chess.MainWindowSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyUp},
	}

	ReduceBoardWidthMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SetDotDistance(Chess.DotCanvasDistance - 10)
			Message.Send("Now BoardWidth: %v", Chess.MainWindowSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyDown},
	}

	ResetBoardWidthMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SetDotDistance(DefaultDotDistance)
			Message.Send("Now BoardWidth: %v", Chess.MainWindowSize)
		},
	}

	AIPlayer1MenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.StartAIPlayer1()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key1},
	}

	AIPlayer2MenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.StartAIPlayer2()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key2},
	}

	IncreaseAISearchTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchTime = Chess.AISearchTime << 1
			Message.Send("Now AISearchTime: %v", Chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key3},
	}

	ReduceAISearchTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchTime = Chess.AISearchTime >> 1
			Message.Send("Now AISearchTime: %v", Chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key4},
	}

	ResetAISearchTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchTime = time.Second
			Message.Send("Now AISearchTime: %v", Chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key5},
	}

	IncreaseSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchGoroutines <<= 1
			Message.Send("Now AISearchGoroutines: %v", Chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key6},
	}

	ReduceSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchGoroutines >>= 1
			Message.Send("Now AISearchGoroutines: %v", Chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key7},
	}

	ResetSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Chess.AISearchGoroutines = runtime.NumCPU()
			Message.Send("Now AISearchGoroutines: %v", Chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key8},
	}

	MusicMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			Message.Send(GetMessage("Music", !Chess.OpenMusic))
			Chess.OpenMusic = !Chess.OpenMusic
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyP},
	}

	AutoRestartMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			if !Chess.AutoRestartGame {
				if CurrentBoard.Size() == AllEdgesCount {
					game.Restart(Chess.BoardSize)
				}
			}
			message := GetMessage("Auto Restart Game", !Chess.AutoRestartGame)
			Message.Send(message)
			Chess.AutoRestartGame = !Chess.AutoRestartGame
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyA},
	}

	QuitMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			Message.Send("Game Closed")
			game.Refresh()
			os.Exit(0)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyQ},
		IsQuit:   true,
	}

	var performanceAnalysisLock sync.Mutex // Mutex for performance analysis synchronization

	IncreasePerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			Chess.PerformanceAnalysisTime += 5 * time.Second
			Message.Send("Now PerformanceAnalysisTime: %v", Chess.PerformanceAnalysisTime)
		},
	}

	ReducePerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			Chess.PerformanceAnalysisTime -= 5 * time.Second
			Message.Send("Now PerformanceAnalysisTime: %v", Chess.PerformanceAnalysisTime)
		},
	}

	ResetPerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			Chess.PerformanceAnalysisTime = DefaultPerformanceAnalysisTime
			Message.Send("Now PerformanceAnalysisTime: %v", Chess.PerformanceAnalysisTime)
		},
	}

	SavePerformanceAnalysisMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			SavePerformanceAnalysisMenuItem.Disabled = true
			MainWindow.MainMenu().Refresh()
			defer func() {
				SavePerformanceAnalysisMenuItem.Disabled = false
				MainWindow.MainMenu().Refresh()
			}()
			r := gin.Default()
			ginpprof.Register(r)
			srv := &http.Server{Addr: ":6060", Handler: r}
			defer func(srv *http.Server) {
				if err := srv.Close(); err != nil {
					Message.Send(err.Error())
				}
			}(srv)
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					Message.Send(err.Error())
				}
			}()
			Message.Send("Start to generate pprof")
			pprofFileName := fmt.Sprintf("%v-%v.prof", time.Now().Format(time.DateTime), time.Now().Add(Chess.PerformanceAnalysisTime).Format(time.DateTime))
			f, err := os.Create(pprofFileName)
			if err != nil {
				Message.Send(err.Error())
				return
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				Message.Send(err.Error())
				return
			}
			time.Sleep(Chess.PerformanceAnalysisTime)
			pprof.StopCPUProfile()
			Message.Send("Finish to generate Performance Analysis: %v", pprofFileName)
			if err := f.Close(); err != nil {
				Message.Send(err.Error())
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyF},
	}

	HelpMenuItem = &fyne.MenuItem{
		Action: func() {
			link, err := url.Parse(HelpDocUrl)
			if err != nil {
				Message.Send(err.Error())
			}
			if err := fyne.CurrentApp().OpenURL(link); err != nil {
				Message.Send(err.Error())
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyH},
	}

	MainWindow.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu(
				"Game",
				RestartGameMenuItem,
				UndoMenuItem,
				ScoreMenuItem,
				SaveScreenshotMenuItem,
				QuitMenuItem,
				fyne.NewMenuItemSeparator(),
				HelpMenuItem,
			),
			fyne.NewMenu(
				"Board",
				IncreaseBoardWidthMenuItem,
				ReduceBoardWidthMenuItem,
				ResetBoardWidthMenuItem,
				fyne.NewMenuItemSeparator(),
				IncreaseBoardSizeMenuItem,
				ReduceBoardSizeMenuItem,
				ResetBoardSizeMenuItem,
			),
			fyne.NewMenu(
				"Config",
				AIPlayer1MenuItem,
				AIPlayer2MenuItem,
				fyne.NewMenuItemSeparator(),
				IncreaseAISearchTimeMenuItem,
				ReduceAISearchTimeMenuItem,
				ResetAISearchTimeMenuItem,
				IncreaseSearchGoroutinesMenuItem,
				ReduceSearchGoroutinesMenuItem,
				ResetSearchGoroutinesMenuItem,
				fyne.NewMenuItemSeparator(),
				AutoRestartMenuItem,
				MusicMenuItem,
			),
			fyne.NewMenu(
				"Performance Analysis",
				IncreasePerformanceAnalysisTimeMenuItem,
				ReducePerformanceAnalysisTimeMenuItem,
				ResetPerformanceAnalysisTimeMenuItem,
				SavePerformanceAnalysisMenuItem,
			),
		),
	)
}

func RefreshMenu() {
	RestartGameMenuItem.Disabled = CurrentBoard.Size() == 0
	RestartGameMenuItem.Label = "Restart"

	MusicMenuItem.Disabled = false
	MusicMenuItem.Label = GetMessage("Music", !Chess.OpenMusic)

	AIPlayer1MenuItem.Disabled = false
	AIPlayer1MenuItem.Label = GetMessage("AIPlayer1", !Chess.AIPlayer1)

	AIPlayer2MenuItem.Disabled = false
	AIPlayer2MenuItem.Label = GetMessage("AIPlayer2", !Chess.AIPlayer2)

	AutoRestartMenuItem.Disabled = false
	AutoRestartMenuItem.Label = GetMessage("AutoRestart", !Chess.AutoRestartGame)

	IncreaseBoardWidthMenuItem.Disabled = false
	IncreaseBoardWidthMenuItem.Label = "Add BoardWidth"

	ReduceBoardWidthMenuItem.Disabled = Chess.DotCanvasDistance <= MinDotSize
	ReduceBoardWidthMenuItem.Label = "Reduce BoardWidth"

	ResetBoardWidthMenuItem.Disabled = Chess.DotCanvasDistance == DefaultDotDistance
	ResetBoardWidthMenuItem.Label = "Reset BoardWidth"

	IncreaseBoardSizeMenuItem.Disabled = false
	IncreaseBoardSizeMenuItem.Label = "Add BoardSize"

	ReduceBoardSizeMenuItem.Disabled = Chess.BoardSize <= MinBoardSize
	ReduceBoardSizeMenuItem.Label = "Reduce BoardSize"

	ResetBoardSizeMenuItem.Disabled = Chess.BoardSize == DefaultBoardSize
	ResetBoardSizeMenuItem.Label = "Reset BoardSize"

	QuitMenuItem.Disabled = false
	QuitMenuItem.Label = "Quit"

	UndoMenuItem.Disabled = CurrentBoard.Size() == 0
	UndoMenuItem.Label = "Undo"

	ScoreMenuItem.Disabled = false
	ScoreMenuItem.Label = "Score"

	IncreaseAISearchTimeMenuItem.Disabled = false
	IncreaseAISearchTimeMenuItem.Label = "Increase AI Search Time"

	ReduceAISearchTimeMenuItem.Disabled = Chess.AISearchTime < time.Millisecond
	ReduceAISearchTimeMenuItem.Label = "Reduce AI Search Time"

	ResetAISearchTimeMenuItem.Disabled = Chess.AISearchTime == DefaultStepTime
	ResetAISearchTimeMenuItem.Label = "Reset AI Search Time"

	IncreaseSearchGoroutinesMenuItem.Disabled = false
	IncreaseSearchGoroutinesMenuItem.Label = "Increase Search Goroutines"

	ReduceSearchGoroutinesMenuItem.Disabled = Chess.AISearchGoroutines <= 1
	ReduceSearchGoroutinesMenuItem.Label = "Reduce Search Goroutines"

	ResetSearchGoroutinesMenuItem.Disabled = Chess.AISearchGoroutines == runtime.NumCPU()
	ResetSearchGoroutinesMenuItem.Label = "Reset Search Goroutines"

	IncreasePerformanceAnalysisTimeMenuItem.Disabled = false
	IncreasePerformanceAnalysisTimeMenuItem.Label = "Increase Performance Analysis Time"

	ReducePerformanceAnalysisTimeMenuItem.Disabled = Chess.PerformanceAnalysisTime <= time.Second*5
	ReducePerformanceAnalysisTimeMenuItem.Label = "Reduce Performance Analysis Time"

	ResetPerformanceAnalysisTimeMenuItem.Disabled = Chess.PerformanceAnalysisTime == DefaultPerformanceAnalysisTime
	ResetPerformanceAnalysisTimeMenuItem.Label = "Reset Performance Analysis Time"

	SavePerformanceAnalysisMenuItem.Disabled = false
	SavePerformanceAnalysisMenuItem.Label = "Save CPU Performance Analysis"

	HelpMenuItem.Disabled = false
	HelpMenuItem.Label = "Help"

	SaveScreenshotMenuItem.Disabled = false
	SaveScreenshotMenuItem.Label = "Save Screenshot"

	MainWindow.MainMenu().Refresh()
}
