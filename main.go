package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/HuXin0817/colog"
	"github.com/bytedance/sonic"
	"image/color"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	BoardSize      = 6
	BoardSizePower = BoardSize * BoardSize

	AIPlayer1 = true
	AIPlayer2 = true

	DotDistance    = float32(75)
	DotWidth       = float32(15)
	DotMargin      = float32(50)
	BoxSize        = DotDistance - DotWidth
	MainWindowSize = DotDistance*float32(BoardSize) + DotMargin

	Goroutines = 32
	SearchTime = 1e7
)

type Turn int

const (
	Player1Turn Turn = 1
	Player2Turn Turn = -1
)

func (t *Turn) Change() { *t = -*t }

type (
	Dot   int
	Box   int
	Edge  int
	Board map[Edge]struct{}

	AssessData struct {
		SumScore   float64 `json:"s"`
		SearchTime float64 `json:"t"`
	}
)

var (
	GlobalBoard = make(Board)

	EdgeFilledColor       = color.RGBA{R: 128, G: 128, B: 128, A: 128}
	Player1HighLightColor = color.NRGBA{R: 30, G: 30, B: 255, A: 128}
	Player2HighLightColor = color.NRGBA{R: 255, G: 30, B: 30, A: 128}
	TipColor              = color.NRGBA{R: 255, G: 255, B: 30, A: 50}
	Player1FilledColor    = color.NRGBA{R: 30, G: 30, B: 128, A: 128}
	Player2FilledColor    = color.NRGBA{R: 128, G: 30, B: 30, A: 128}

	signChan = make(chan struct{}, 1)
	mu       sync.Mutex

	NowTurn      = Player1Turn
	Player1Score = 0
	Player2Score = 0
	AssessTable  = make(map[string]map[Edge]AssessData)
	AssessFile   = "assess_table.json"

	EdgesUICanvases  = make(map[Edge]*canvas.Line)
	BoxesUICanvases  = make(map[Box]*canvas.Rectangle)
	Buttons          = make(map[Edge]*widget.Button)
	BoxesFilledColor = make(map[Box]color.Color)
	Container        *fyne.Container
	MainWindow       = app.New().NewWindow("Dots and Boxes")
)

func init() {
	if len(os.Args) == 2 {
		AssessFile = os.Args[1]
	}

	file, err := os.ReadFile(AssessFile)
	if err != nil {
		return
	}

	if err = sonic.Unmarshal(file, &AssessTable); err != nil {
		panic(err)
	}
}

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

var Dots = func() (Dots []Dot) {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			Dots = append(Dots, NewDot(i, j))
		}
	}
	return
}()

func NewEdge(Dot1, Dot2 Dot) Edge {
	if Dot1 > Dot2 {
		Dot1, Dot2 = Dot2, Dot1
	}
	return Edge(Dot1*Dot(BoardSizePower) + Dot2)
}

func (e Edge) Dot1() Dot { return Dot(e) / Dot(BoardSizePower) }

func (e Edge) Dot2() Dot { return Dot(e) % Dot(BoardSizePower) }

func (e Edge) ToString() string {
	return fmt.Sprintf("(%d, %d) => (%d, %d)", e.Dot1().X(), e.Dot1().Y(), e.Dot2().X(), e.Dot2().Y())
}

var EdgeNearBoxes = func() map[Edge][]Box {
	edges := make(map[Edge][]Box)
	for e := range EdgesMap {
		x := e.Dot2().X() - 1
		y := e.Dot2().Y() - 1
		if x >= 0 && y >= 0 {
			boxes := []Box{Box(e.Dot1()), Box(NewDot(x, y))}
			edges[e] = boxes
			continue
		}
		boxes := []Box{Box(e.Dot1())}
		edges[e] = boxes
	}
	return edges
}()

func (e Edge) NearBoxes() []Box { return EdgeNearBoxes[e] }

var Edges = func() (Edges []Edge) {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			d := NewDot(i, j)
			if i+1 < BoardSize {
				Edges = append(Edges, NewEdge(d, NewDot(i+1, j)))
			}
			if j+1 < BoardSize {
				Edges = append(Edges, NewEdge(d, NewDot(i, j+1)))
			}
		}
	}
	return
}()

var EdgesMap = func() map[Edge]struct{} {
	EdgesMap := make(map[Edge]struct{})
	for _, e := range Edges {
		EdgesMap[e] = struct{}{}
	}
	return EdgesMap
}()

var BoxEdges = func() map[Box][]Edge {
	BoxEdges := make(map[Box][]Edge)
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
			NewEdge(D10, D11),
			NewEdge(D01, D11),
		}

		BoxEdges[b] = edges
	}
	return BoxEdges
}()

func (b Box) Edges() []Edge { return BoxEdges[b] }

var Boxes = func() (Boxes []Box) {
	for _, d := range Dots {
		if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
			Boxes = append(Boxes, Box(d))
		}
	}
	return Boxes
}()

func NewBoard(board Board) Board {
	b := make(Board, len(board))
	for e := range board {
		b[e] = struct{}{}
	}
	return b
}

func (b Board) ToString() (s string) {
	for _, e := range Edges {
		if _, c := b[e]; c {
			s += "1"
		} else {
			s += "0"
		}
	}
	return
}

func (b Board) EdgesCountInBox(box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b[e]; c {
			count++
		}
	}
	return
}

func (b Board) ObtainsScore(e Edge) (count int) {
	if _, c := b[e]; c {
		return
	}

	boxes := e.NearBoxes()
	for _, box := range boxes {
		if b.EdgesCountInBox(box) == 3 {
			count++
		}
	}
	return
}

func (b Board) ObtainsBoxes(e Edge) (obtainsBoxes []Box) {
	if _, c := b[e]; c {
		return
	}

	boxes := e.NearBoxes()
	for _, box := range boxes {
		if b.EdgesCountInBox(box) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}
	return
}

func transPosition(x int) float32 { return DotMargin + float32(x)*DotDistance }

func getDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

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

func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8((float64(r1)*(1-t) + float64(r2)*t) / 256)
	g := uint8((float64(g1)*(1-t) + float64(g2)*t) / 256)
	b := uint8((float64(b1)*(1-t) + float64(b2)*t) / 256)
	a := uint8((float64(a1)*(1-t) + float64(a2)*t) / 256)

	return color.RGBA{R: r, G: g, B: b, A: a}
}

func AddEdge(e Edge) {
	mu.Lock()
	defer mu.Unlock()

	defer Container.Refresh()
	if NowTurn == Player1Turn {
		EdgesUICanvases[e].StrokeColor = Player1HighLightColor
	} else {
		EdgesUICanvases[e].StrokeColor = Player2HighLightColor
	}

	score := GlobalBoard.ObtainsScore(e)
	if NowTurn == Player1Turn {
		Player1Score += score
	} else {
		Player2Score += score
	}

	boxes := GlobalBoard.ObtainsBoxes(e)
	for _, box := range boxes {
		if NowTurn == Player1Turn {
			BoxesUICanvases[box].FillColor = Player1FilledColor
			BoxesFilledColor[box] = Player1FilledColor
		} else {
			BoxesUICanvases[box].FillColor = Player2FilledColor
			BoxesFilledColor[box] = Player2FilledColor
		}
	}

	if NowTurn == Player1Turn {
		colog.Infof("Step: %d Player1 Edge: %s Player1 Score: %d, Player2 Score: %d", len(GlobalBoard), e.ToString(), Player1Score, Player2Score)
	} else {
		colog.Infof("Step: %d Player2 Edge: %s Player1 Score: %d, Player2 Score: %d", len(GlobalBoard), e.ToString(), Player1Score, Player2Score)
	}

	if button, c := Buttons[e]; c {
		button.Hide()
	}

	if score == 0 {
		NowTurn.Change()
	}

	GlobalBoard[e] = struct{}{}

	for _, box := range Boxes {
		edgesCountInBox := GlobalBoard.EdgesCountInBox(box)
		if edgesCountInBox == 3 {
			go func() {
				startColor := TipColor
				endColor := color.Black
				step := 100
				d := time.Second / time.Duration(step)
				nowTurn := len(GlobalBoard)

				for {
					for i := 0; i <= step; i++ {
						if nowTurn != len(GlobalBoard) {
							BoxesUICanvases[box].FillColor = BoxesFilledColor[box]
							BoxesUICanvases[box].Refresh()
							return
						}

						time.Sleep(d)
						t := float64(i) / float64(step)
						BoxesUICanvases[box].FillColor = interpolateColor(startColor, endColor, t)
						BoxesUICanvases[box].Refresh()
					}
					for i := 0; i <= step; i++ {
						if nowTurn != len(GlobalBoard) {
							BoxesUICanvases[box].FillColor = BoxesFilledColor[box]
							BoxesUICanvases[box].Refresh()
							return
						}

						time.Sleep(d)
						t := float64(i) / float64(step)
						BoxesUICanvases[box].FillColor = interpolateColor(endColor, startColor, t)
						BoxesUICanvases[box].Refresh()
					}
				}
			}()
		}
	}

	if len(EdgesMap) == len(GlobalBoard) {
		switch {
		case Player1Score > Player2Score:
			colog.Info("Player1 Win! Score:", Player1Score)
		case Player1Score < Player2Score:
			colog.Info("Player2 Win! Score:", Player2Score)
		case Player1Score == Player2Score:
			colog.Infof("Draw!")
		}

		time.Sleep(time.Second * 2)
		os.Exit(0)
	}

	if AIPlayer1 && NowTurn == Player1Turn {
		signChan <- struct{}{}
	} else if AIPlayer2 && NowTurn == Player2Turn {
		signChan <- struct{}{}
	}
}

func GetNextEdges(board Board) Edge {
	minEnemyCanGetScore := 3
	var minEnemyCanGetEdge Edge

	for e := range EdgesMap {
		if _, c := board[e]; !c {
			if score := board.ObtainsScore(e); score > 0 {
				return e
			} else if score == 0 {
				boxes := e.NearBoxes()
				s := 0
				for _, box := range boxes {
					if board.EdgesCountInBox(box) == 2 {
						s++
					}
				}
				if minEnemyCanGetScore > s {
					minEnemyCanGetScore = s
					minEnemyCanGetEdge = e
				}
			}
		}
	}

	return minEnemyCanGetEdge
}

func GetBestEdge() (bestEdge Edge) {
	var mutex sync.Mutex
	boardStr := GlobalBoard.ToString()
	if _, c := AssessTable[boardStr]; !c {
		AssessTable[boardStr] = make(map[Edge]AssessData)
	}

	assessDataTable := AssessTable[boardStr]

	sonsSumSearchTime := 0
	for _, v := range assessDataTable {
		sonsSumSearchTime += int(v.SearchTime)
	}

	freeEdgeCount := len(Edges) - len(GlobalBoard)
	searchTime := SearchTime / freeEdgeCount
	if sonsSumSearchTime >= searchTime {
		searchTime /= 1e3
	}

	var t atomic.Int64
	var wg sync.WaitGroup
	wg.Add(Goroutines)
	for range Goroutines {
		go func() {
			defer wg.Done()
			for int(t.Load()) < searchTime {
				b := NewBoard(GlobalBoard)
				turn := Player1Turn
				var firstEdge Edge
				score := 0.0

				for len(b) < len(EdgesMap) {
					t.Add(1)
					edge := GetNextEdges(b)
					if firstEdge == 0 {
						firstEdge = edge
					}
					s := b.ObtainsScore(edge)
					score += float64(turn) * float64(s)
					if s == 0 {
						turn.Change()
					}
					b[edge] = struct{}{}
				}

				mutex.Lock()
				assessDataTable[firstEdge] = AssessData{
					SumScore:   assessDataTable[firstEdge].SumScore + score,
					SearchTime: assessDataTable[firstEdge].SearchTime + 1,
				}
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	bestScore := -1e9
	for e, data := range assessDataTable {
		avgScore := data.SumScore / data.SearchTime
		if avgScore > bestScore {
			bestEdge = e
			bestScore = avgScore
		}
	}

	AssessTable[boardStr] = assessDataTable
	j, err := sonic.Marshal(AssessTable)
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(AssessFile, j, 0644); err != nil {
		panic(err)
	}

	return
}

func main() {
	if err := colog.OpenLog("gamelog/" + time.Now().Format(time.DateTime) + ".log"); err != nil {
		panic(err)
	}

	background := canvas.NewRectangle(color.Black)
	background.Move(fyne.NewPos(0, 0))
	background.Resize(fyne.NewSize(1e10, 1e10))

	Container = container.NewWithoutLayout(background)

	for _, b := range Boxes {
		boxUi := NewBox(b)
		BoxesUICanvases[b] = boxUi
		Container.Add(boxUi)
		BoxesFilledColor[b] = color.Black
	}

	for e := range EdgesMap {
		edgeUi := NewEdgeCanvas(e)
		EdgesUICanvases[e] = edgeUi
		Container.Add(edgeUi)
		Buttons[e] = widget.NewButton("", func() {
			switch {
			case AIPlayer1 && NowTurn == Player1Turn:
				return
			case AIPlayer2 && NowTurn == Player2Turn:
				return
			}
			AddEdge(e)
		})

		var SizeX, SizeY float32
		if e.Dot1().X() == e.Dot2().X() {
			SizeX = DotWidth
			SizeY = DotDistance
		} else {
			SizeX = DotDistance
			SizeY = DotWidth
		}

		Buttons[e].Resize(fyne.NewSize(SizeX, SizeY))
		PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - Buttons[e].Size().Width/2 + DotWidth/2
		PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - Buttons[e].Size().Height/2 + DotWidth/2
		Buttons[e].Move(fyne.NewPos(PosX, PosY))
		Container.Add(Buttons[e])
	}

	for _, d := range Dots {
		dotUi := NewDotCanvas(d)
		Container.Add(dotUi)
	}

	go func() {
		for range signChan {
			if AIPlayer1 && NowTurn == Player1Turn {
				AddEdge(GetBestEdge())
			} else if AIPlayer2 && NowTurn == Player2Turn {
				AddEdge(GetBestEdge())
			}
		}
	}()

	MainWindow.Resize(fyne.NewSize(MainWindowSize, MainWindowSize))
	MainWindow.SetContent(Container)
	MainWindow.SetFixedSize(true)

	if AIPlayer1 {
		go func() {
			time.Sleep(time.Second >> 1)
			AddEdge(GetBestEdge())
		}()
	}

	colog.Info("GAME START!")
	MainWindow.ShowAndRun()
}
