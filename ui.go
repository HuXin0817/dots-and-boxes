package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"sync"
	"time"
)

var (
	DotDistance = float32(75)
	DotWidth    = float32(15)
	DotMargin   = float32(50)

	EdgeFilledColor       = color.RGBA{R: 128, G: 128, B: 128, A: 128}
	Player1HighLightColor = color.NRGBA{R: 30, G: 30, B: 255, A: 128}
	Player2HighLightColor = color.NRGBA{R: 255, G: 30, B: 30, A: 128}

	Player1FilledColor = color.NRGBA{R: 30, G: 30, B: 128, A: 128}
	Player2FilledColor = color.NRGBA{R: 128, G: 30, B: 30, A: 128}

	BoxSize = DotDistance - DotWidth

	mainWindowSize = DotDistance*float32(BoardSize) + DotMargin
)

func transPosition(x int) float32 {
	return DotMargin + float32(x)*DotDistance
}

func getDotPosition(d Dot) (float32, float32) {
	return transPosition(d.X()), transPosition(d.Y())
}

func NewDotCanvas(d Dot) (newDotCanvas *canvas.Circle) {
	newDotCanvas = canvas.NewCircle(color.White)
	newDotCanvas.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotCanvas.Move(fyne.NewPos(getDotPosition(d)))
	return
}

func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + DotWidth/2
	x2 := transPosition(e.Dot2().X()) + DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + DotWidth/2

	newEdge := canvas.NewLine(EdgeFilledColor)
	newEdge.Position1 = fyne.NewPos(x1, y1)
	newEdge.Position2 = fyne.NewPos(x2, y2)
	newEdge.StrokeWidth = DotWidth
	return newEdge
}

func NewBox(s Box) *canvas.Rectangle {
	d := Dot(s)
	x := transPosition(d.X()) + DotWidth
	y := transPosition(d.Y()) + DotWidth

	r := canvas.NewRectangle(color.Black)
	r.Move(fyne.NewPos(x, y))
	r.Resize(fyne.NewSize(BoxSize, BoxSize))
	return r
}

type BoardUI struct {
	board        Board
	aiPlayer1    bool
	aiPlayer2    bool
	player1Score int
	player2Score int
	NowTurn      Turn
	Container    *fyne.Container
	edges        map[Edge]*canvas.Line
	boxes        map[Box]*canvas.Rectangle
	buttons      map[Edge]*widget.Button
	signChan     chan struct{}
	mu           sync.Mutex
}

func NewBoardUI() (board *BoardUI) {
	background := canvas.NewRectangle(color.Black)
	background.Move(fyne.NewPos(0, 0))
	background.Resize(fyne.NewSize(1e10, 1e10))

	board = &BoardUI{
		board:     make(Board),
		NowTurn:   Player1Turn,
		Container: container.NewWithoutLayout(background),
		edges:     make(map[Edge]*canvas.Line),
		boxes:     make(map[Box]*canvas.Rectangle),
		buttons:   make(map[Edge]*widget.Button),
		signChan:  make(chan struct{}, 1),
	}

	for _, b := range Boxes {
		boxUi := NewBox(b)
		board.boxes[b] = boxUi
		board.Container.Add(boxUi)
	}

	for _, e := range EdgesMap {
		edgeUi := NewEdgeCanvas(e)
		board.edges[e] = edgeUi
		board.Container.Add(edgeUi)
		board.buttons[e] = widget.NewButton("", func() {
			switch {
			case AIPlayer1 && board.NowTurn == Player1Turn:
				return
			case AIPlayer2 && board.NowTurn == Player2Turn:
				return
			}
			board.AddEdge(e)
		})

		var SizeX, SizeY float32
		if e.Dot1().X() == e.Dot2().X() {
			SizeX = DotWidth
			SizeY = DotDistance
		} else {
			SizeX = DotDistance
			SizeY = DotWidth
		}

		board.buttons[e].Resize(fyne.NewSize(SizeX, SizeY))
		PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - board.buttons[e].Size().Width/2 + DotWidth/2
		PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - board.buttons[e].Size().Height/2 + DotWidth/2
		board.buttons[e].Move(fyne.NewPos(PosX, PosY))
		board.Container.Add(board.buttons[e])
	}

	for _, d := range Dots {
		dotUi := NewDotCanvas(d)
		board.Container.Add(dotUi)
	}

	go func() {
		for range board.signChan {
			e := GenerateBestEdge(board.board)
			board.AddEdge(e)
		}
	}()

	return
}

func (b *BoardUI) AddEdge(e Edge) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, c := b.board[e]; c {
		return
	}

	defer b.Container.Refresh()

	if button, c := b.buttons[e]; c {
		button.Hide()
	} else {
		return
	}

	player := b.NowTurn

	if player == Player1Turn {
		b.edges[e].StrokeColor = Player1HighLightColor
	} else {
		b.edges[e].StrokeColor = Player2HighLightColor
	}

	boxes := b.board.ObtainsBoxes(e)
	for _, box := range boxes {
		if player == Player1Turn {
			b.boxes[box].FillColor = Player1FilledColor
		} else {
			b.boxes[box].FillColor = Player2FilledColor
		}
	}

	score := b.board.ObtainsScore(e)
	if b.NowTurn == Player1Turn {
		b.player1Score += score
	} else {
		b.player2Score += score
	}

	if score == 0 {
		b.NowTurn.Change()
	}

	if player == Player1Turn {
		log.Printf("Player1 Move, Edge: %s\n", e.ToString())
	}

	if player == Player2Turn {
		log.Printf("Player2 Move, Edge: %s\n", e.ToString())
	}

	b.board[e] = struct{}{}
	if b.player1Score+len(EdgesMap)-len(b.board) < b.player2Score || b.player2Score+len(EdgesMap)-len(b.board) < b.player1Score || len(EdgesMap) == len(b.board) {
		switch {
		case b.player1Score > b.player2Score:
			fmt.Println("Player1 Win! Score:", b.player1Score)
		case b.player2Score > b.player1Score:
			fmt.Println("Player2 Win! Score:", b.player2Score)
		case b.player1Score == b.player2Score:
			fmt.Println("Draw!")
		}

		time.Sleep(time.Second * 3)
		os.Exit(0)
	}

	fmt.Printf("Player1 Score: %d, Player2 Score: %d\n\n", b.player1Score, b.player2Score)

	if b.aiPlayer1 && b.NowTurn == Player1Turn {
		b.signChan <- struct{}{}
	} else if b.aiPlayer2 && b.NowTurn == Player2Turn {
		b.signChan <- struct{}{}
	}
}
