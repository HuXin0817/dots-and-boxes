package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var (
	RestartMenuItem          *fyne.MenuItem
	MusicMenuItem            *fyne.MenuItem
	AIPlayer1MenuItem        *fyne.MenuItem
	AIPlayer2MenuItem        *fyne.MenuItem
	AutoRestartMenuItem      *fyne.MenuItem
	PauseMenuItem            *fyne.MenuItem
	AddBoardSizeMenuItem     *fyne.MenuItem
	ReduceBoardSizeMenuItem  *fyne.MenuItem
	UndoMenuItem             *fyne.MenuItem
	AddBoardWidthMenuItem    *fyne.MenuItem
	ReduceBoardWidthMenuItem *fyne.MenuItem
	ResetBoardMenuItem       *fyne.MenuItem
	ScoreMenuItem            *fyne.MenuItem
	QuitMenuItem             *fyne.MenuItem
	HelpMenuItem             *fyne.MenuItem
)

var (
	PauseState               bool
	TmpMenuItemDisabledState = make(map[string]bool)
	UnDisableMenuItemLabel   = map[string]struct{}{
		"Continue":  {},
		"Quit":      {},
		"Help":      {},
		"Music ON":  {},
		"Music OFF": {},
	}
)

func init() {
	RestartMenuItem = &fyne.MenuItem{
		Label: "Restart",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			defer Refresh()
			RestartWithCall(BoardSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	PauseMenuItem = &fyne.MenuItem{
		Label:    "Pause",
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeySpace},
		Action: func() {
			defer MainWindow.MainMenu().Refresh()
			if !PauseState {
				mu.Lock()
				PauseMenuItem.Label = "Continue"
				SendMessage("Game Paused")
				TmpMenuItemDisabledState = make(map[string]bool)
				for _, menu := range MainWindow.MainMenu().Items {
					for _, m := range menu.Items {
						if _, c := UnDisableMenuItemLabel[m.Label]; !c {
							TmpMenuItemDisabledState[m.Label] = m.Disabled
							m.Disabled = true
						}
					}
				}
			} else {
				for _, menu := range MainWindow.MainMenu().Items {
					for _, m := range menu.Items {
						if _, c := UnDisableMenuItemLabel[m.Label]; !c {
							m.Disabled = TmpMenuItemDisabledState[m.Label]
						}
					}
				}
				mu.Unlock()
				PauseMenuItem.Label = "Pause"
				SendMessage("Game Continue")
			}
			PauseState = !PauseState
		},
	}

	ScoreMenuItem = &fyne.MenuItem{
		Label: "Score",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			SendMessage("Player1 Score: %d\nPlayer2 Score: %d\n", PlayerScore[Player1Turn], PlayerScore[Player2Turn])
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS},
	}

	AddBoardSizeMenuItem = &fyne.MenuItem{
		Label: "AddBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			RestartWithCall(BoardSize + 1)
			if !ReduceBoardSizeMenuItem.Disabled {
				ReduceBoardSizeMenuItem.Disabled = false
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyEqual},
	}

	ReduceBoardSizeMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			defer Refresh()
			if BoardSize <= 1 {
				return
			}
			RestartWithCall(BoardSize - 1)
			if BoardSize <= 1 {
				ResetBoardMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyMinus},
	}

	UndoMenuItem = &fyne.MenuItem{
		Label: "Undo",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			moveRecord := append([]Edge{}, MoveRecord...)
			if len(moveRecord) > 0 {
				e := moveRecord[len(moveRecord)-1]
				SendMessage("Undo Edge " + e.ToString())
				moveRecord = moveRecord[:len(moveRecord)-1]
				Recover(moveRecord)
			}
			if len(moveRecord) == 0 {
				UndoMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Disabled: true,
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyZ},
	}

	AddBoardWidthMenuItem = &fyne.MenuItem{
		Label: "AddBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			SetDotDistance(DotDistance + 10)
			if ReduceBoardWidthMenuItem.Disabled {
				ReduceBoardWidthMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyUp},
	}

	ReduceBoardWidthMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			if DotDistance-10 < MinDotSize {
				return
			}
			SetDotDistance(DotDistance - 10)
			if DotDistance-10 < MinDotSize {
				ReduceBoardSizeMenuItem.Disabled = true
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyDown},
	}

	ResetBoardMenuItem = &fyne.MenuItem{
		Label: "ResetBoard",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			ResetBoard()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	AIPlayer1MenuItem = &fyne.MenuItem{
		Label: "AIPlayer1 ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AIPlayer1.Value() {
				defer notifySignChan()
				AIPlayer1MenuItem.Label = "AIPlayer1 OFF"
			} else {
				AIPlayer1MenuItem.Label = "AIPlayer1 ON"
			}
			AIPlayer1.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key1},
	}

	AIPlayer2MenuItem = &fyne.MenuItem{
		Label: "AIPlayer2 ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AIPlayer2.Value() {
				defer notifySignChan()
				AIPlayer2MenuItem.Label = "AIPlayer2 OFF"
			} else {
				AIPlayer2MenuItem.Label = "AIPlayer2 ON"
			}
			AIPlayer2.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key2},
	}

	MusicMenuItem = &fyne.MenuItem{
		Label: "Music OFF",
		Action: func() {
			defer MainWindow.MainMenu().Refresh()
			if !MusicOn.Value() {
				MusicMenuItem.Label = "Music OFF"
			} else {
				MusicMenuItem.Label = "Music ON"
			}
			MusicOn.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyP},
	}

	AutoRestartMenuItem = &fyne.MenuItem{
		Label: "AutoRestart ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AutoRestart.Value() {
				AutoRestartMenuItem.Label = "AutoRestart OFF"
			} else {
				AutoRestartMenuItem.Label = "AutoRestart ON"
			}
			AutoRestart.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyA},
	}

	QuitMenuItem = &fyne.MenuItem{
		Label: "Quit",
		Action: func() {
			SendMessage("Game Closed")
			os.Exit(0)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyQ},
		IsQuit:   true,
	}

	HelpMenuItem = &fyne.MenuItem{
		Label: "Help",
		Action: func() {
			SendMessage(HelpDoc)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyH},
	}

	MainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Game",
			RestartMenuItem,
			UndoMenuItem,
			PauseMenuItem,
			ScoreMenuItem,
			fyne.NewMenuItemSeparator(),
			QuitMenuItem,
		),
		fyne.NewMenu("Board",
			AddBoardWidthMenuItem,
			ReduceBoardWidthMenuItem,
			AddBoardSizeMenuItem,
			ReduceBoardSizeMenuItem,
			fyne.NewMenuItemSeparator(),
			ResetBoardMenuItem,
		),
		fyne.NewMenu("Config",
			AIPlayer1MenuItem,
			AIPlayer2MenuItem,
			AutoRestartMenuItem,
			MusicMenuItem,
		),
		fyne.NewMenu("Help",
			HelpMenuItem,
		),
	))
}
