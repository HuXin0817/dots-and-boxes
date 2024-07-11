package main

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func SetDotDistance(d float32) {
	DotDistance = d
	DotWidth = DotDistance / 5
	DotMargin = DotDistance / 3 * 2
	BoxSize = DotDistance - DotWidth
	MainWindowSize = DotDistance*float32(BoardSize) + DotMargin - 5
	MainWindow.Resize(fyne.NewSize(MainWindowSize, MainWindowSize))
	moveRecord := append([]Edge{}, MoveRecord...)
	Recover(moveRecord)
}

func transPosition(x int) float32 { return DotMargin + float32(x)*DotDistance }

func GetDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

func getEdgeButtonSizeAndPosition(e Edge) (size fyne.Size, pos fyne.Position) {
	if e.Dot1().X() == e.Dot2().X() {
		size = fyne.NewSize(DotWidth, DotDistance)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+DotWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+DotWidth/2,
		)
	} else {
		size = fyne.NewSize(DotDistance, DotWidth)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+DotWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+DotWidth/2,
		)
	}
	return
}

func NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(GetDotCanvasColor())
	newDotCanvas.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotCanvas.Move(fyne.NewPos(GetDotPosition(d)))
	return newDotCanvas
}

func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + DotWidth/2
	x2 := transPosition(e.Dot2().X()) + DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + DotWidth/2
	newEdgeCanvas := canvas.NewLine(GetDotCanvasColor())
	newEdgeCanvas.Position1 = fyne.NewPos(x1, y1)
	newEdgeCanvas.Position2 = fyne.NewPos(x2, y2)
	newEdgeCanvas.StrokeWidth = DotWidth
	return newEdgeCanvas
}

func NewBoxCanvas(box Box) *canvas.Rectangle {
	d := Dot(box)
	x := transPosition(d.X()) + DotWidth
	y := transPosition(d.Y()) + DotWidth
	newBoxCanvas := canvas.NewRectangle(GetThemeColor())
	newBoxCanvas.Move(fyne.NewPos(x, y))
	newBoxCanvas.Resize(fyne.NewSize(BoxSize, BoxSize))
	return newBoxCanvas
}

func Refresh() {
	defer Container.Refresh()
	img := MainWindow.Canvas().Capture()
	file, err := os.Create(ImagePath)
	if err != nil {
		return
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return
	}
	MainWindow.SetIcon(fyne.NewStaticResource("Dots-and-Boxes", buf.Bytes()))
}

func notifySignChan() {
	if AIPlayer1.Value() && NowTurn == Player1Turn || AIPlayer2.Value() && NowTurn == Player2Turn {
		select {
		case SignChan <- struct{}{}:
		default:
		}
	}
}

func NewGame(boardSize int) {
	BoardSize = boardSize
	BoardSizePower = Dot(BoardSize * BoardSize)
	MainWindowSize = DotDistance*float32(BoardSize) + DotMargin - 5
	MainWindow.Resize(fyne.NewSize(MainWindowSize, MainWindowSize))
	Dots = []Dot{}
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			Dots = append(Dots, NewDot(i, j))
		}
	}
	FullBoard = make(Board)
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			d := NewDot(i, j)
			if i+1 < BoardSize {
				FullBoard[NewEdge(d, NewDot(i+1, j))] = struct{}{}
			}
			if j+1 < BoardSize {
				FullBoard[NewEdge(d, NewDot(i, j+1))] = struct{}{}
			}
		}
	}
	EdgesCount = len(FullBoard)
	Boxes = []Box{}
	for _, d := range Dots {
		if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
			Boxes = append(Boxes, Box(d))
		}
	}
	EdgeNearBoxes = make(map[Edge][]Box)
	for e := range FullBoard {
		x := e.Dot2().X() - 1
		y := e.Dot2().Y() - 1
		if x >= 0 && y >= 0 {
			EdgeNearBoxes[e] = []Box{Box(e.Dot1()), Box(NewDot(x, y))}
			continue
		}
		EdgeNearBoxes[e] = []Box{Box(e.Dot1())}
	}
	BoxEdges = make(map[Box][]Edge)
	for _, b := range Boxes {
		x := Dot(b).X()
		y := Dot(b).Y()
		D00 := NewDot(x, y)
		D10 := NewDot(x+1, y)
		D01 := NewDot(x, y+1)
		D11 := NewDot(x+1, y+1)
		edges := []Edge{
			NewEdge(D00, D01),
			NewEdge(D00, D10),
			NewEdge(D01, D11),
			NewEdge(D10, D11),
		}
		BoxEdges[b] = edges
	}
	DotCanvases = make(map[Dot]*canvas.Circle)
	EdgesCanvases = make(map[Edge]*canvas.Line)
	BoxesCanvases = make(map[Box]*canvas.Rectangle)
	EdgeButtons = make(map[Edge]*widget.Button)
	BoxesFilledColor = make(map[Box]color.Color)
	Container = container.NewWithoutLayout()
	SignChan = make(chan struct{}, 1)
	MoveRecord = []Edge{}
	NowTurn = Player1Turn
	PlayerScore = map[Turn]int{Player1Turn: 0, Player2Turn: 0}
	GlobalBoard = make(Board)
	for _, b := range Boxes {
		BoxesCanvases[b] = NewBoxCanvas(b)
		Container.Add(BoxesCanvases[b])
	}
	for e := range FullBoard {
		EdgesCanvases[e] = NewEdgeCanvas(e)
		Container.Add(EdgesCanvases[e])
		EdgeButtons[e] = widget.NewButton("", func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if AIPlayer1.Value() && NowTurn == Player1Turn {
				return
			} else if AIPlayer2.Value() && NowTurn == Player2Turn {
				return
			}
			AddEdge(e)
		})
		size, pos := getEdgeButtonSizeAndPosition(e)
		EdgeButtons[e].Resize(size)
		EdgeButtons[e].Move(pos)
		Container.Add(EdgeButtons[e])
	}
	for _, d := range Dots {
		DotCanvases[d] = NewDotCanvas(d)
		Container.Add(DotCanvases[d])
	}
	go func() {
		notifySignChan()
		for range SignChan {
			mu.Lock()
			AddEdge(GetBestEdge())
			Refresh()
			mu.Unlock()
		}
	}()
	MainWindow.SetContent(Container)
}

func RestartWithCall(size int) {
	NewGame(size)
	SendMessage("Game Start! BoardSize: %d", BoardSize)
}

func Tip(nowStep int, box Box) {
	boxesCanvas := BoxesCanvases[box]
	defer func() {
		boxesCanvasMu.Lock()
		defer boxesCanvasMu.Unlock()
		boxesCanvas.FillColor = BoxesFilledColor[box]
	}()
	ticker := time.NewTicker(AnimationStepTime)
	defer ticker.Stop()
	for {
		for i := 0; i <= AnimationSteps; i++ {
			if nowStep != len(GlobalBoard) {
				return
			}
			t := float64(i) / float64(AnimationSteps)
			boxesCanvas.FillColor = interpolateColor(TipColor, GetThemeColor(), t)
			boxesCanvas.Refresh()
			<-ticker.C
		}
		for i := 0; i <= AnimationSteps; i++ {
			if nowStep != len(GlobalBoard) {
				return
			}
			t := float64(i) / float64(AnimationSteps)
			boxesCanvas.FillColor = interpolateColor(GetThemeColor(), TipColor, t)
			boxesCanvas.Refresh()
			<-ticker.C
		}
	}
}

func AddEdge(e Edge) {
	if BoardSize <= 1 {
		return
	}
	if _, c := GlobalBoard[e]; c && e > 0 {
		return
	}
	defer func() { MoveRecord = append(MoveRecord, e) }()
	nowStep := len(GlobalBoard)
	obtainsBoxes := ObtainsBoxes(GlobalBoard, e)
	score := len(obtainsBoxes)
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		if score > 0 {
			PlayScoreMusic()
		} else {
			PlayMoveMusic()
		}
	}()
	defer EdgeButtons[e].Hide()
	boxesCanvasMu.Lock()
	for _, box := range obtainsBoxes {
		playerFilledColor := GetPlayerFilledColor()
		BoxesCanvases[box].FillColor = playerFilledColor
		BoxesFilledColor[box] = playerFilledColor
	}
	boxesCanvasMu.Unlock()
	EdgesCanvases[e].StrokeColor = GetPlayerHighlightColor()
	PlayerScore[NowTurn] += score
	if score == 0 {
		NowTurn.Change()
	}
	GlobalBoard[e] = struct{}{}
	nowStep++
	for _, box := range Boxes {
		edgesCountInBox := EdgesCountInBox(GlobalBoard, box)
		if edgesCountInBox == 3 {
			go Tip(nowStep, box)
		}
	}
	if nowStep == EdgesCount {
		if PlayerScore[Player1Turn] > PlayerScore[Player2Turn] {
			SendMessage("Player1 Win!")
		} else if PlayerScore[Player1Turn] < PlayerScore[Player2Turn] {
			SendMessage("Player2 Win!")
		} else if PlayerScore[Player1Turn] == PlayerScore[Player2Turn] {
			SendMessage("Draw!")
		}
		if AutoRestart.Value() {
			go func() {
				time.Sleep(2 * time.Second)
				RestartWithCall(BoardSize)
			}()
		}
		return
	}
	notifySignChan()
	UndoMenuItem.Disabled = false
}

func Recover(MoveRecord []Edge) {
	if MusicOn.Value() {
		MusicOn.Change()
		defer MusicOn.Change()
	}
	NewGame(BoardSize)
	for _, e := range MoveRecord {
		AddEdge(e)
	}
}

func ResetBoard() {
	if DotDistance != DefaultDotDistance {
		SetDotDistance(DefaultDotDistance)
	}
	if BoardSize != DefaultDotSize {
		RestartWithCall(DefaultDotSize)
	}
	ResetBoardMenuItem.Disabled = true
}
