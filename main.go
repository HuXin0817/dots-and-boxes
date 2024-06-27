package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"time"
)

const (
	BoardSize = 6
	AIPlayer1 = true
	AIPlayer2 = true
)

var mainWindow = app.New().NewWindow("Dots and boxes")

func main() {
	board := NewBoardUI()
	board.aiPlayer1 = AIPlayer1
	board.aiPlayer2 = AIPlayer2

	mainWindow.Resize(fyne.NewSize(mainWindowSize, mainWindowSize))
	mainWindow.SetContent(board.Container)

	if board.aiPlayer1 {
		go func() {
			time.Sleep(time.Second)
			e := GenerateBestEdge(board.board)
			board.AddEdge(e)
		}()
	}

	mainWindow.ShowAndRun()
}
