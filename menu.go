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
	switch runtime.GOOS {
	case "darwin":
		return getSaveFilePathMac()
	case "windows":
		return getSaveFilePathWindows()
	case "linux":
		return getSaveFilePathLinux()
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

// getSaveFilePathMac prompts the user to choose a file path on macOS.
func getSaveFilePathMac() (string, error) {
	script := `osascript -e 'set myFile to choose file name with prompt "Save game screenshot as:" default name "dots-and-boxes screenshot.png"' -e 'POSIX path of myFile'`
	cmd := exec.Command("sh", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getSaveFilePathWindows prompts the user to choose a file path on Windows.
func getSaveFilePathWindows() (string, error) {
	// Implement Windows specific file save dialog using PowerShell or other methods
	script := `Add-Type -AssemblyName System.Windows.Forms; $file = New-Object System.Windows.Forms.SaveFileDialog; $file.Filter = "PNG Files|*.png"; $file.FileName = "dots-and-boxes screenshot.png"; if($file.ShowDialog() -eq 'OK') {$file.FileName}`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getSaveFilePathLinux prompts the user to choose a file path on Linux.
func getSaveFilePathLinux() (string, error) {
	// Implement Linux specific file save dialog using zenity, kdialog, or other methods
	script := `zenity --file-selection --save --confirm-overwrite --file-filter="*.png" --filename="dots-and-boxes screenshot.png"`
	cmd := exec.Command("sh", "-c", script)
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
			game.Restart(chess.BoardSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	ScoreMenuItem = &fyne.MenuItem{
		Action: func() {
			message := fmt.Sprintf("Player1 Score: %v\nPlayer2 Score: %v\n", Player1Score, Player2Score)
			SendMessage(message)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyT},
	}

	IncreaseBoardSizeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(chess.BoardSize + 1)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyEqual},
	}

	ReduceBoardSizeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			game.Restart(chess.BoardSize - 1)
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
				SendMessage(err.Error())
				return
			}
			path, err := getSaveFilePath()
			if err != nil {
				SendMessage(err.Error())
				return
			}
			if err := os.WriteFile(path, buf.Bytes(), 0666); err != nil {
				SendMessage(err.Error())
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
			SetDotDistance(chess.DotCanvasDistance + 10)
			SendMessage("Now BoardWidth: %v", chess.MainWindowSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyUp},
	}

	ReduceBoardWidthMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SetDotDistance(chess.DotCanvasDistance - 10)
			SendMessage("Now BoardWidth: %v", chess.MainWindowSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyDown},
	}

	ResetBoardWidthMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SetDotDistance(DefaultDotDistance)
			SendMessage("Now BoardWidth: %v", chess.MainWindowSize)
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
			chess.AISearchTime = chess.AISearchTime << 1
			SendMessage("Now AISearchTime: %v", chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key3},
	}

	ReduceAISearchTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			chess.AISearchTime = chess.AISearchTime >> 1
			SendMessage("Now AISearchTime: %v", chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key4},
	}

	ResetAISearchTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			chess.AISearchTime = time.Second
			SendMessage("Now AISearchTime: %v", chess.AISearchTime)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key5},
	}

	IncreaseSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			chess.AISearchGoroutines <<= 1
			SendMessage("Now AISearchGoroutines: %v", chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key6},
	}

	ReduceSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			chess.AISearchGoroutines >>= 1
			SendMessage("Now AISearchGoroutines: %v", chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key7},
	}

	ResetSearchGoroutinesMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			chess.AISearchGoroutines = runtime.NumCPU()
			SendMessage("Now AISearchGoroutines: %v", chess.AISearchGoroutines)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key8},
	}

	MusicMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			SendMessage(GetMessage("Music", !chess.OpenMusic))
			chess.OpenMusic = !chess.OpenMusic
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyP},
	}

	AutoRestartMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			defer game.Refresh()
			if !chess.AutoRestartGame {
				if CurrentBoard.Size() == AllEdgesCount {
					game.Restart(chess.BoardSize)
				}
			}
			message := GetMessage("Auto Restart Game", !chess.AutoRestartGame)
			SendMessage(message)
			chess.AutoRestartGame = !chess.AutoRestartGame
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyA},
	}

	QuitMenuItem = &fyne.MenuItem{
		Action: func() {
			globalLock.Lock()
			SendMessage("Game Closed")
			game.Refresh()
			os.Exit(0)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyQ},
		IsQuit:   true,
	}

	IncreasePerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			chess.PerformanceAnalysisTime += 5 * time.Second
			SendMessage("Now PerformanceAnalysisTime: %v", chess.PerformanceAnalysisTime)
		},
	}

	ReducePerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			chess.PerformanceAnalysisTime -= 5 * time.Second
			SendMessage("Now PerformanceAnalysisTime: %v", chess.PerformanceAnalysisTime)
		},
	}

	ResetPerformanceAnalysisTimeMenuItem = &fyne.MenuItem{
		Action: func() {
			performanceAnalysisLock.Lock()
			defer performanceAnalysisLock.Unlock()
			chess.PerformanceAnalysisTime = DefaultPerformanceAnalysisTime
			SendMessage("Now PerformanceAnalysisTime: %v", chess.PerformanceAnalysisTime)
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
					SendMessage(err.Error())
				}
			}(srv)
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					SendMessage(err.Error())
				}
			}()
			SendMessage("Start to generate pprof")
			pprofFileName := fmt.Sprintf("%v-%v.prof", time.Now().Format(time.DateTime), time.Now().Add(chess.PerformanceAnalysisTime).Format(time.DateTime))
			f, err := os.Create(pprofFileName)
			if err != nil {
				SendMessage(err.Error())
				return
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				SendMessage(err.Error())
				return
			}
			time.Sleep(chess.PerformanceAnalysisTime)
			pprof.StopCPUProfile()
			SendMessage("Finish to generate Performance Analysis: %v", pprofFileName)
			if err := f.Close(); err != nil {
				SendMessage(err.Error())
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyF},
	}

	HelpMenuItem = &fyne.MenuItem{
		Action: func() {
			link, err := url.Parse(HelpDocUrl)
			if err != nil {
				SendMessage(err.Error())
			}
			if err := fyne.CurrentApp().OpenURL(link); err != nil {
				SendMessage(err.Error())
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
	MusicMenuItem.Label = GetMessage("Music", !chess.OpenMusic)

	AIPlayer1MenuItem.Disabled = false
	AIPlayer1MenuItem.Label = GetMessage("AIPlayer1", !chess.AIPlayer1)

	AIPlayer2MenuItem.Disabled = false
	AIPlayer2MenuItem.Label = GetMessage("AIPlayer2", !chess.AIPlayer2)

	AutoRestartMenuItem.Disabled = false
	AutoRestartMenuItem.Label = GetMessage("AutoRestart", !chess.AutoRestartGame)

	IncreaseBoardWidthMenuItem.Disabled = false
	IncreaseBoardWidthMenuItem.Label = "Add BoardWidth"

	ReduceBoardWidthMenuItem.Disabled = chess.DotCanvasDistance <= MinDotSize
	ReduceBoardWidthMenuItem.Label = "Reduce BoardWidth"

	ResetBoardWidthMenuItem.Disabled = chess.DotCanvasDistance == DefaultDotDistance
	ResetBoardWidthMenuItem.Label = "Reset BoardWidth"

	IncreaseBoardSizeMenuItem.Disabled = false
	IncreaseBoardSizeMenuItem.Label = "Add BoardSize"

	ReduceBoardSizeMenuItem.Disabled = chess.BoardSize <= MinBoardSize
	ReduceBoardSizeMenuItem.Label = "Reduce BoardSize"

	ResetBoardSizeMenuItem.Disabled = chess.BoardSize == DefaultBoardSize
	ResetBoardSizeMenuItem.Label = "Reset BoardSize"

	QuitMenuItem.Disabled = false
	QuitMenuItem.Label = "Quit"

	UndoMenuItem.Disabled = CurrentBoard.Size() == 0
	UndoMenuItem.Label = "Undo"

	ScoreMenuItem.Disabled = false
	ScoreMenuItem.Label = "Score"

	IncreaseAISearchTimeMenuItem.Disabled = false
	IncreaseAISearchTimeMenuItem.Label = "Increase AI Search Time"

	ReduceAISearchTimeMenuItem.Disabled = chess.AISearchTime < time.Millisecond
	ReduceAISearchTimeMenuItem.Label = "Reduce AI Search Time"

	ResetAISearchTimeMenuItem.Disabled = chess.AISearchTime == DefaultStepTime
	ResetAISearchTimeMenuItem.Label = "Reset AI Search Time"

	IncreaseSearchGoroutinesMenuItem.Disabled = false
	IncreaseSearchGoroutinesMenuItem.Label = "Increase Search Goroutines"

	ReduceSearchGoroutinesMenuItem.Disabled = chess.AISearchGoroutines <= 1
	ReduceSearchGoroutinesMenuItem.Label = "Reduce Search Goroutines"

	ResetSearchGoroutinesMenuItem.Disabled = chess.AISearchGoroutines == runtime.NumCPU()
	ResetSearchGoroutinesMenuItem.Label = "Reset Search Goroutines"

	IncreasePerformanceAnalysisTimeMenuItem.Disabled = false
	IncreasePerformanceAnalysisTimeMenuItem.Label = "Increase Performance Analysis Time"

	ReducePerformanceAnalysisTimeMenuItem.Disabled = chess.PerformanceAnalysisTime <= time.Second*5
	ReducePerformanceAnalysisTimeMenuItem.Label = "Reduce Performance Analysis Time"

	ResetPerformanceAnalysisTimeMenuItem.Disabled = chess.PerformanceAnalysisTime == DefaultPerformanceAnalysisTime
	ResetPerformanceAnalysisTimeMenuItem.Label = "Reset Performance Analysis Time"

	SavePerformanceAnalysisMenuItem.Disabled = false
	SavePerformanceAnalysisMenuItem.Label = "Save CPU Performance Analysis"

	HelpMenuItem.Disabled = false
	HelpMenuItem.Label = "Help"

	SaveScreenshotMenuItem.Disabled = false
	SaveScreenshotMenuItem.Label = "Save Screenshot"

	MainWindow.MainMenu().Refresh()
}
