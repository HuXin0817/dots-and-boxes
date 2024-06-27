package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"sync"
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

	mainWindow = app.New().NewWindow("Dots and Boxes")
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
	GameInformation     *Game
	AIPlayer1           bool
	AIPlayer2           bool
	Container           *fyne.Container
	Edges               map[Edge]*canvas.Line
	Boxes               map[Box]*canvas.Rectangle
	Dots                map[Dot]*canvas.Circle
	Buttons             map[Edge]*widget.Button
	TriggerAfterAddEdge func()
	LastChangedEdge     Edge
	mu                  sync.Mutex
}

func NewBoardUI(BoardSize int) (newBoard *BoardUI) {
	background := canvas.NewRectangle(color.Black)
	background.Move(fyne.NewPos(0, 0))
	background.Resize(fyne.NewSize(1e10, 1e10))

	newBoard = &BoardUI{
		GameInformation:     NewGame(BoardSize),
		Container:           container.NewWithoutLayout(background),
		Edges:               make(map[Edge]*canvas.Line),
		Boxes:               make(map[Box]*canvas.Rectangle),
		Dots:                make(map[Dot]*canvas.Circle),
		Buttons:             make(map[Edge]*widget.Button),
		TriggerAfterAddEdge: func() {},
	}

	boxes := Boxes(BoardSize)
	for _, b := range boxes {
		boxUi := NewBox(b)
		newBoard.Boxes[b] = boxUi
		newBoard.Container.Add(boxUi)
	}

	edges := Edges(BoardSize)
	for _, e := range edges {
		edgeUi := NewEdgeCanvas(e)
		newBoard.Edges[e] = edgeUi
		newBoard.Container.Add(edgeUi)
		newBoard.Buttons[e] = widget.NewButton("", func() {
			switch {
			case bool(newBoard.AIPlayer1) && newBoard.GameInformation.NowPlayer == Player1:
				return
			case bool(newBoard.AIPlayer2) && newBoard.GameInformation.NowPlayer == Player2:
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
		PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - newBoard.Buttons[e].Size().Width/2 + DotWidth/2
		PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - newBoard.Buttons[e].Size().Height/2 + DotWidth/2
		newBoard.Buttons[e].Move(fyne.NewPos(PosX, PosY))
		newBoard.Container.Add(newBoard.Buttons[e])
	}

	dots := Dots(BoardSize)
	for _, d := range dots {
		dotUi := NewDotCanvas(d)
		newBoard.Dots[d] = dotUi
		newBoard.Container.Add(dotUi)
	}

	return
}

func (b *BoardUI) AddEdge(e Edge) {
	defer b.TriggerAfterAddEdge()

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, c := b.GameInformation.Board.Edges[e]; c {
		return
	}

	b.Buttons[e].Hide()
	player := b.GameInformation.NowPlayer

	if player == Player1 {
		b.Edges[e].StrokeColor = Player1HighLightColor
	}

	if player == Player2 {
		b.Edges[e].StrokeColor = Player2HighLightColor
	}

	boxes := b.GameInformation.Board.ObtainsBoxes(e)
	for _, box := range boxes {
		if player == Player1 {
			b.Boxes[box].FillColor = Player1FilledColor
		}

		if player == Player2 {
			b.Boxes[box].FillColor = Player2FilledColor
		}
	}

	b.GameInformation.Add(e)
	b.Container.Refresh()

	if player == Player1 {
		log.Printf("Player1 Move, Edge: %s\n", e.String())
	}

	if player == Player2 {
		log.Printf("Player2 Move, Edge: %s\n", e.String())
	}

	fmt.Printf("Player1 Score: %d, Player2 Score: %d\n\n", b.GameInformation.Player1Score, b.GameInformation.Player2Score)
	b.LastChangedEdge = e

	if b.GameInformation.FreeEdgesCount() == 0 {
		switch {
		case b.GameInformation.Player1Score > b.GameInformation.Player2Score:
			fmt.Println("Player1 Win!, Score:", b.GameInformation.Player1Score)
		case b.GameInformation.Player2Score > b.GameInformation.Player1Score:
			fmt.Println("Player2 Win!, Score:", b.GameInformation.Player2Score)
		case b.GameInformation.Player1Score == b.GameInformation.Player2Score:
			fmt.Println("Draw!")
		}

		b.Container.Refresh()
	}
}

func (b *BoardUI) Run() {
	size := DotDistance*float32(b.GameInformation.BoardSize) + DotMargin
	mainWindow.Resize(fyne.NewSize(size, size))
	mainWindow.SetContent(b.Container)
	mainWindow.ShowAndRun()
}
