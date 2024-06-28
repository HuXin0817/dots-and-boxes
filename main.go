package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/HuXin0817/colog"
	"time"
)

const (
	BoardSize = 6
	AIPlayer1 = true
	AIPlayer2 = true
)

var mainWindow = app.New().NewWindow("Dots and boxes")

func main() {
	err := colog.OpenLog("gamelog/" + time.Now().Format(time.DateTime) + ".log")
	if err != nil {
		panic(err)
	}

	board := NewBoardUI()
	board.aiPlayer1 = AIPlayer1
	board.aiPlayer2 = AIPlayer2

	mainWindow.Resize(fyne.NewSize(mainWindowSize, mainWindowSize))
	mainWindow.SetContent(board.Container)
	mainWindow.SetFixedSize(true)

	if board.aiPlayer1 {
		go func() {
			time.Sleep(time.Second)
			e := GenerateBestEdge(board.board)
			board.AddEdge(e)
		}()
	}

	mainWindow.ShowAndRun()
}
