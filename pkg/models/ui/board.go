package ui

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"sync"
	"time"

	"github.com/HuXin0817/dots-and-boxes/bgm"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/music"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Board struct {
	GameInformation     *chess.Game
	AIPlayer1           model.Config
	AIPlayer2           model.Config
	Container           *fyne.Container
	Edges               map[chess.Edge]*canvas.Line
	Boxes               map[chess.Box]*canvas.Rectangle
	Dots                map[chess.Dot]*canvas.Circle
	Buttons             map[chess.Edge]*widget.Button
	TriggerAfterAddEdge func()
	LastChangedEdge     chess.Edge
	mu                  sync.Mutex
}

func NewBoard(BoardSize int) (newBoard *Board) {
	background := canvas.NewRectangle(color.Black)
	background.Move(fyne.NewPos(0, 0))
	background.Resize(fyne.NewSize(1e10, 1e10))

	newBoard = &Board{
		GameInformation:     chess.NewGame(BoardSize),
		Container:           container.NewWithoutLayout(background),
		Edges:               make(map[chess.Edge]*canvas.Line),
		Boxes:               make(map[chess.Box]*canvas.Rectangle),
		Dots:                make(map[chess.Dot]*canvas.Circle),
		Buttons:             make(map[chess.Edge]*widget.Button),
		TriggerAfterAddEdge: func() {},
	}

	boxes := chess.Boxes(BoardSize)
	for _, b := range boxes {
		boxUi := NewBox(b)
		newBoard.Boxes[b] = boxUi
		newBoard.Container.Add(boxUi)
	}

	edges := chess.Edges(BoardSize)
	for _, e := range edges {
		edgeUi := NewEdge(e)
		newBoard.Edges[e] = edgeUi
		newBoard.Container.Add(edgeUi)
		newBoard.Buttons[e] = widget.NewButton("", func() {
			switch {
			case bool(newBoard.AIPlayer1) && newBoard.GameInformation.NowPlayer == chess.Player1:
				return
			case bool(newBoard.AIPlayer2) && newBoard.GameInformation.NowPlayer == chess.Player2:
				return
			}

			newBoard.AddEdge(e)
		})

		var SizeX, SizeY float32
		if e.Dot1().X() == e.Dot2().X() {
			SizeX = DotWidth
			SizeY = DotDistance
		} else {
			SizeX = DotDistance
			SizeY = DotWidth
		}

		newBoard.Buttons[e].Resize(fyne.NewSize(SizeX, SizeY))
		PosX := (getPosition(e.Dot1().X())+getPosition(e.Dot2().X()))/2 - newBoard.Buttons[e].Size().Width/2 + DotWidth/2
		PosY := (getPosition(e.Dot1().Y())+getPosition(e.Dot2().Y()))/2 - newBoard.Buttons[e].Size().Height/2 + DotWidth/2
		newBoard.Buttons[e].Move(fyne.NewPos(PosX, PosY))
		newBoard.Container.Add(newBoard.Buttons[e])
	}

	dots := chess.Dots(BoardSize)
	for _, d := range dots {
		dotUi := NewDot(d)
		newBoard.Dots[d] = dotUi
		newBoard.Container.Add(dotUi)
	}

	return
}

func (b *Board) AddEdge(e chess.Edge) {
	defer b.TriggerAfterAddEdge()

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, c := b.GameInformation.Board.Edges[e]; c {
		return
	}

	b.Buttons[e].Hide()
	player := b.GameInformation.NowPlayer

	if player == chess.Player1 {
		b.Edges[e].StrokeColor = Player1HighLightColor
	}

	if player == chess.Player2 {
		b.Edges[e].StrokeColor = Player2HighLightColor
	}

	boxes := b.GameInformation.Board.ObtainsBoxes(e)
	for _, box := range boxes {
		if player == chess.Player1 {
			b.Boxes[box].FillColor = Player1Color
		}

		if player == chess.Player2 {
			b.Boxes[box].FillColor = Player2Color
		}
	}

	music.BeepPlay(bgm.Touchmp3)
	switch len(boxes) {
	case 1:
		music.BeepPlay(bgm.Obtain_one_score_mp3)
	case 2:
		music.BeepPlay(bgm.Obtain_two_score_mp3)
	}

	b.GameInformation.Add(e)
	b.Container.Refresh()

	if player == chess.Player1 {
		log.Printf("Player 1 Move, Edge: %s\n", e.String())
	}

	if player == chess.Player2 {
		log.Printf("Player 2 Move, Edge: %s\n", e.String())
	}

	fmt.Printf("Player 1 Score: %d, Player 2 Score: %d\n\n", b.GameInformation.Player1Score, b.GameInformation.Player2Score)
	b.LastChangedEdge = e

	if b.GameInformation.FreeEdgesCount() == 0 {
		go music.BeepPlay(bgm.Gameover_mp3)

		switch {
		case b.GameInformation.Player1Score > b.GameInformation.Player2Score:
			fmt.Println("Player 1 Win!, Score:", b.GameInformation.Player1Score)
		case b.GameInformation.Player2Score > b.GameInformation.Player1Score:
			fmt.Println("Player 2 Win!, Score:", b.GameInformation.Player2Score)
		case b.GameInformation.Player1Score == b.GameInformation.Player2Score:
			fmt.Println("Draw!")
		}

		allBoxes := chess.Boxes(b.GameInformation.BoardSize)
		for _, box := range allBoxes {
			image := NewImage(chess.Dot(box).X(), chess.Dot(box).Y())
			b.Container.Add(image)
		}

		b.Container.Refresh()
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
}

var mainWindow = app.New().NewWindow("Dots and Boxes")

func (b *Board) Run() {
	size := DotDistance*float32(b.GameInformation.BoardSize) + DotMargin
	mainWindow.Resize(fyne.NewSize(size, size))
	mainWindow.SetContent(b.Container)
	mainWindow.ShowAndRun()
}
