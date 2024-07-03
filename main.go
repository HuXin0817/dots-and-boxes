package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/HuXin0817/colog"
)

type (
	Turn       int
	Dot        int
	Box        int
	Edge       int
	Board      map[Edge]struct{}
	AssessData struct {
		SumScore   float64 `json:"s"`
		SearchTime float64 `json:"t"`
	}
)

const (
	Player1Turn Turn = 1
	Player2Turn Turn = -1
)

const (
	BoardSize      = 6
	BoardSizePower = BoardSize * BoardSize
	AIPlayer1      = true
	AIPlayer2      = true
	DotDistance    = 75
	DotWidth       = 15
	DotMargin      = 50
	BoxSize        = DotDistance - DotWidth
	MainWindowSize = DotDistance*BoardSize + DotMargin
	SearchTime     = 1e6
	Goroutines     = 32
)

var (
	HighLightColor = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 255, A: 128},
		Player2Turn: {R: 255, G: 30, B: 30, A: 128},
	}
	FilledColor = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 128, A: 128},
		Player2Turn: {R: 128, G: 30, B: 30, A: 128},
	}
	TipColor         = color.NRGBA{R: 255, G: 255, B: 30, A: 50}
	EdgesUICanvases  = make(map[Edge]*canvas.Line)
	BoxesUICanvases  = make(map[Box]*canvas.Rectangle)
	Buttons          = make(map[Edge]*widget.Button)
	BoxesFilledColor = make(map[Box]color.Color)
	Container        *fyne.Container
	MainWindow       = app.New().NewWindow("Dots and Boxes")
	SignChan         = make(chan struct{}, 1)
	NowTurn          = Player1Turn
	PlayerScore      = map[Turn]int{Player1Turn: 0, Player2Turn: 0}
	AssessTable      = make(map[string]map[Edge]AssessData)
	AssessFile       = "assess.json"
	GlobalBoard      = make(Board)
	mu               sync.Mutex

	Dots = func() (Dots []Dot) {
		for i := 0; i < BoardSize; i++ {
			for j := 0; j < BoardSize; j++ {
				Dots = append(Dots, NewDot(i, j))
			}
		}
		return
	}()

	EdgesCount = len(Edges)

	Edges = func() (Edges []Edge) {
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

	EdgesMap = func() map[Edge]struct{} {
		EdgesMap := make(map[Edge]struct{})
		for _, e := range Edges {
			EdgesMap[e] = struct{}{}
		}
		return EdgesMap
	}()

	Boxes = func() (Boxes []Box) {
		for _, d := range Dots {
			if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
				Boxes = append(Boxes, Box(d))
			}
		}
		return Boxes
	}()

	EdgeNearBoxes = func() map[Edge][]Box {
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

	BoxEdges = func() map[Box][]Edge {
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
				NewEdge(D01, D11),
				NewEdge(D10, D11),
			}
			BoxEdges[b] = edges
		}
		return BoxEdges
	}()

	TurnToString = map[Turn]string{
		Player1Turn: "Player1",
		Player2Turn: "Player2",
	}
)

func (t *Turn) ToString() string { return TurnToString[*t] }

func (t *Turn) Change() { *t = -*t }

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*BoardSizePower + Dot2) }

func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

func (e Edge) ToString() string {
	return fmt.Sprintf("(%d, %d) => (%d, %d)", e.Dot1().X(), e.Dot1().Y(), e.Dot2().X(), e.Dot2().Y())
}

func (e Edge) NearBoxes() []Box { return EdgeNearBoxes[e] }

func (b Box) Edges() []Edge { return BoxEdges[b] }

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

func NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(color.White)
	newDotCanvas.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotCanvas.Move(fyne.NewPos(getDotPosition(d)))
	return newDotCanvas
}

func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + DotWidth/2
	x2 := transPosition(e.Dot2().X()) + DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + DotWidth/2
	newEdge := canvas.NewLine(color.Black)
	newEdge.Position1 = fyne.NewPos(x1, y1)
	newEdge.Position2 = fyne.NewPos(x2, y2)
	newEdge.StrokeWidth = DotWidth
	return newEdge
}

func NewBox(box Box) *canvas.Rectangle {
	d := Dot(box)
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
	defer Buttons[e].Hide()
	nowStep := len(GlobalBoard)
	boxes := GlobalBoard.ObtainsBoxes(e)
	score := len(boxes)
	for _, box := range boxes {
		BoxesUICanvases[box].FillColor = FilledColor[NowTurn]
		BoxesFilledColor[box] = FilledColor[NowTurn]
	}
	EdgesUICanvases[e].StrokeColor = HighLightColor[NowTurn]
	PlayerScore[NowTurn] += score
	colog.Infof("Step: %d, Turn %s, Edge: %s, Player1 Score: %d, Player2 Score: %d", nowStep, NowTurn.ToString(), e.ToString(), PlayerScore[Player1Turn], PlayerScore[Player2Turn])
	if score == 0 {
		NowTurn.Change()
	}
	GlobalBoard[e] = struct{}{}
	nowStep++
	for _, box := range Boxes {
		edgesCountInBox := GlobalBoard.EdgesCountInBox(box)
		if edgesCountInBox == 3 {
			go func() {
				defer func() {
					BoxesUICanvases[box].FillColor = BoxesFilledColor[box]
					BoxesUICanvases[box].Refresh()
				}()
				startColor := TipColor
				endColor := color.Black
				step := 100
				d := time.Second / time.Duration(step)
				for {
					for i := 0; i <= step; i++ {
						if nowStep != len(GlobalBoard) {
							return
						}
						time.Sleep(d)
						t := float64(i) / float64(step)
						BoxesUICanvases[box].FillColor = interpolateColor(startColor, endColor, t)
						BoxesUICanvases[box].Refresh()
					}
					for i := 0; i <= step; i++ {
						if nowStep != len(GlobalBoard) {
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
	if nowStep == EdgesCount {
		timer := time.NewTimer(2 * time.Second)
		if PlayerScore[Player1Turn] > PlayerScore[Player2Turn] {
			colog.Info("Player1 Win!")
		} else if PlayerScore[Player1Turn] < PlayerScore[Player2Turn] {
			colog.Info("Player2 Win!")
		} else if PlayerScore[Player1Turn] == PlayerScore[Player2Turn] {
			colog.Infof("Draw!")
		}
		if j, err := json.Marshal(AssessTable); err == nil {
			os.WriteFile(AssessFile, j, 0644)
		}
		<-timer.C
		os.Exit(0)
	}
	if AIPlayer1 && NowTurn == Player1Turn {
		SignChan <- struct{}{}
	} else if AIPlayer2 && NowTurn == Player2Turn {
		SignChan <- struct{}{}
	}
}

func GetNextEdges(board Board) (bestEdge Edge) {
	enemyMinScore := 3
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
				if enemyMinScore > s {
					enemyMinScore = s
					bestEdge = e
				}
			}
		}
	}
	return
}

func GetBestEdge() (bestEdge Edge) {
	boardStr := GlobalBoard.ToString()
	if _, c := AssessTable[boardStr]; !c {
		AssessTable[boardStr] = make(map[Edge]AssessData)
	}
	assessDataTable := AssessTable[boardStr]
	var t atomic.Int64
	var wg sync.WaitGroup
	var latch sync.Mutex
	wg.Add(Goroutines)
	for range Goroutines {
		go func() {
			defer wg.Done()
			for t.Load() < SearchTime {
				b := NewBoard(GlobalBoard)
				turn := Player1Turn
				firstEdge := Edge(0)
				score := 0
				for len(b) < EdgesCount {
					t.Add(1)
					edge := GetNextEdges(b)
					if firstEdge == 0 {
						firstEdge = edge
					}
					s := b.ObtainsScore(edge)
					score += int(turn) * s
					if s == 0 {
						turn.Change()
					}
					b[edge] = struct{}{}
				}
				latch.Lock()
				assessDataTable[firstEdge] = AssessData{
					SumScore:   assessDataTable[firstEdge].SumScore + float64(score),
					SearchTime: assessDataTable[firstEdge].SearchTime + 1,
				}
				latch.Unlock()
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
	return
}

func main() {
	if len(os.Args) >= 2 {
		AssessFile = os.Args[1]
	}
	if file, err := os.ReadFile(AssessFile); err == nil {
		json.Unmarshal(file, &AssessTable)
	}
	logFile := filepath.Join("gamelog", time.Now().Format(time.DateTime)+".log")
	if err := colog.OpenLog(logFile); err != nil {
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
			if AIPlayer1 && NowTurn == Player1Turn {
				return
			} else if AIPlayer2 && NowTurn == Player2Turn {
				return
			}
			AddEdge(e)
		})
		if e.Dot1().X() == e.Dot2().X() {
			Buttons[e].Resize(fyne.NewSize(DotWidth, DotDistance))
		} else {
			Buttons[e].Resize(fyne.NewSize(DotDistance, DotWidth))
		}
		PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - Buttons[e].Size().Width/2 + DotWidth/2
		PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - Buttons[e].Size().Height/2 + DotWidth/2
		Buttons[e].Move(fyne.NewPos(PosX, PosY))
		Container.Add(Buttons[e])
	}
	for _, d := range Dots {
		dotUi := NewDotCanvas(d)
		Container.Add(dotUi)
	}
	MainWindow.Resize(fyne.NewSize(MainWindowSize, MainWindowSize))
	MainWindow.SetContent(Container)
	MainWindow.SetFixedSize(true)
	go func() {
		if AIPlayer1 {
			time.Sleep(500 * time.Millisecond)
			AddEdge(GetBestEdge())
		}
		for range SignChan {
			if AIPlayer1 && NowTurn == Player1Turn {
				AddEdge(GetBestEdge())
			} else if AIPlayer2 && NowTurn == Player2Turn {
				AddEdge(GetBestEdge())
			}
		}
	}()
	colog.Info("GAME START!")
	MainWindow.ShowAndRun()
}
