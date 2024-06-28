package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/HuXin0817/colog"
	"time"
)

// 常量定义
const (
	BoardSize = 6    // 棋盘大小
	AIPlayer1 = true // 玩家1是否为AI
	AIPlayer2 = true // 玩家2是否为AI
)

// 创建Fyne应用窗口
var mainWindow = app.New().NewWindow("Dots and boxes")

func main() {
	// 打开日志文件
	if err := colog.OpenLog("gamelog/" + time.Now().Format(time.DateTime) + ".log"); err != nil {
		panic(err)
	}

	// 创建新的棋盘UI
	board := NewBoardUI()
	board.aiPlayer1 = AIPlayer1
	board.aiPlayer2 = AIPlayer2

	// 设置窗口大小和内容
	mainWindow.Resize(fyne.NewSize(mainWindowSize, mainWindowSize))
	mainWindow.SetContent(board.Container)
	mainWindow.SetFixedSize(true) // 固定窗口大小

	if board.aiPlayer1 {
		go func() {
			time.Sleep(time.Second)
			e := GenerateBestEdge(board.board)
			board.AddEdge(e)
		}()
	}

	// 显示窗口并运行应用
	mainWindow.ShowAndRun()
}
