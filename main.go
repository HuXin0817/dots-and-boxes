package main

import "time"

var (
	AIPlayer1 = false
	AIPlayer2 = false
	BoardSize = 6
)

func main() {
	UI := NewBoardUI(BoardSize)
	UI.AIPlayer1 = AIPlayer1
	UI.AIPlayer2 = AIPlayer2

	UI.TriggerAfterAddEdge = func() {
		if UI.AIPlayer1 && UI.GameInformation.NowPlayer == Player1 {
			UI.AddEdge(GetBestMove(UI.GameInformation.Board))
		}

		if UI.AIPlayer2 && UI.GameInformation.NowPlayer == Player2 {
			UI.AddEdge(GetBestMove(UI.GameInformation.Board))
		}
	}

	if UI.AIPlayer1 {
		go func() {
			time.Sleep(time.Second)
			UI.AddEdge(GetBestMove(UI.GameInformation.Board))
		}()
	}

	UI.Run()
}
