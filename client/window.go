package main

import (
	"serve/serveclient"
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/ui"
)

var Window *ui.Board

func initWindow() {
	Window = ui.NewBoard(BoardSize)
	Window.AIPlayer1 = AI1
	Window.AIPlayer2 = AI2
	Window.TriggerAfterAddEdge = addEdgeLogic
}

func GetPostGameInformationRequest() (req *serveclient.GameInformationRequest) {
	edges := make(map[int64]bool)
	for e := range Window.GameInformation.Board.Edges {
		edges[int64(e)] = true
	}

	return &serveclient.GameInformationRequest{
		GameUid:    string(GameUid),
		Timestamp:  string(message.NewTimeStamp(time.Now())),
		BoardSize:  int64(BoardSize),
		StepCount:  int64(len(Window.GameInformation.Board.Edges)),
		Edges:      edges,
		AI1:        bool(AI1),
		AI2:        bool(AI2),
		NowTurn:    Window.GameInformation.NowPlayer.String(),
		Team1Score: int64(Window.GameInformation.Player1Score),
		Team2Score: int64(Window.GameInformation.Player2Score),
		GameOver:   Window.GameInformation.Board.FreeEdgesCount() == 0,
		MoveEdge:   int64(Window.LastChangedEdge),
	}
}
