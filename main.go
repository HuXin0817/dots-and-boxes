package main

import (
	"fmt"
	"image/color"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/HuXin0817/colog"
)

type (
	Turn  int
	Dot   int
	Box   int
	Edge  int
	Board map[Edge]struct{}
)

const (
	Player1Turn Turn = 1
	Player2Turn      = -Player1Turn
)

const (
	BoardSize         = 6
	BoardSizePower    = BoardSize * BoardSize
	AIPlayer1         = true
	AIPlayer2         = true
	DotDistance       = 80
	DotWidth          = DotDistance / 5
	DotMargin         = DotDistance / 3 * 2
	BoxSize           = DotDistance - DotWidth
	MainWindowSize    = DotDistance*BoardSize + DotMargin - 5
	SearchTime        = 1e6
	Goroutines        = 32
	AnimationSteps    = 100
	AnimationStepTime = time.Second / AnimationSteps
	Record            = true
)

var (
	GlobalSystemColor fyne.ThemeVariant
	HighLightColor    = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 255, A: 128},
		Player2Turn: {R: 255, G: 30, B: 30, A: 128},
	}
	FilledColor = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 128, A: 128},
		Player2Turn: {R: 128, G: 30, B: 30, A: 128},
	}
	TipColor = color.NRGBA{R: 255, G: 255, B: 30, A: 50}

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

	EdgesSet = func() map[Edge]struct{} {
		EdgesSet := make(map[Edge]struct{})
		for _, e := range Edges {
			EdgesSet[e] = struct{}{}
		}
		return EdgesSet
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
		EdgeNearBoxes := make(map[Edge][]Box)
		for e := range EdgesSet {
			x := e.Dot2().X() - 1
			y := e.Dot2().Y() - 1
			if x >= 0 && y >= 0 {
				EdgeNearBoxes[e] = []Box{Box(e.Dot1()), Box(NewDot(x, y))}
				continue
			}
			EdgeNearBoxes[e] = []Box{Box(e.Dot1())}
		}
		return EdgeNearBoxes
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
)

var (
	DotCanvases      = make(map[Dot]*canvas.Circle)
	EdgesCanvases    = make(map[Edge]*canvas.Line)
	BoxesCanvases    = make(map[Box]*canvas.Rectangle)
	EdgeButtons      = make(map[Edge]*widget.Button)
	BoxesFilledColor = make(map[Box]color.Color)
	Container        *fyne.Container
	App              = app.New()
	MainWindow       = App.NewWindow("Dots and Boxes")
	SignChan         = make(chan struct{}, 1)
	NowTurn          = Player1Turn
	PlayerScore      = map[Turn]int{Player1Turn: 0, Player2Turn: 0}
	GlobalBoard      = make(Board)
	mu               sync.Mutex
)

func (t *Turn) ToString() string {
	if *t == Player1Turn {
		return "Player1"
	} else {
		return "Player2"
	}
}

func (t *Turn) Change() { *t = -*t }

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

func (d Dot) ToString() string { return fmt.Sprintf("(%d, %d)", d.X(), d.Y()) }

func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*BoardSizePower + Dot2) }

func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

func (e Edge) ToString() string { return e.Dot1().ToString() + " => " + e.Dot2().ToString() }

func (e Edge) NearBoxes() []Box { return EdgeNearBoxes[e] }

func (b Box) Edges() []Edge { return BoxEdges[b] }

func NewBoard(board Board) Board {
	b := make(Board, len(board))
	for e := range board {
		b[e] = struct{}{}
	}
	return b
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
	boxes := e.NearBoxes()
	for _, box := range boxes {
		if b.EdgesCountInBox(box) == 3 {
			count++
		}
	}
	return
}

func (b Board) ObtainsBoxes(e Edge) (obtainsBoxes []Box) {
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

func getDotCanvasColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 202, G: 202, B: 202, A: 255}
	} else {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}

func getThemeColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 43, G: 43, B: 43, A: 255}
	} else {
		return color.RGBA{R: 242, G: 242, B: 242, A: 255}
	}
}

func getButtonColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 65, G: 65, B: 65, A: 255}
	} else {
		return color.RGBA{R: 217, G: 217, B: 217, A: 255}
	}
}

func NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(getDotCanvasColor())
	newDotCanvas.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotCanvas.Move(fyne.NewPos(getDotPosition(d)))
	return newDotCanvas
}

func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + DotWidth/2
	x2 := transPosition(e.Dot2().X()) + DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + DotWidth/2
	newEdgeCanvas := canvas.NewLine(getDotCanvasColor())
	newEdgeCanvas.Position1 = fyne.NewPos(x1, y1)
	newEdgeCanvas.Position2 = fyne.NewPos(x2, y2)
	newEdgeCanvas.StrokeWidth = DotWidth
	return newEdgeCanvas
}

func NewBoxCanvas(box Box) *canvas.Rectangle {
	d := Dot(box)
	x := transPosition(d.X()) + DotWidth - 0.5
	y := transPosition(d.Y()) + DotWidth - 0.5
	newBoxCanvas := canvas.NewRectangle(getThemeColor())
	newBoxCanvas.Move(fyne.NewPos(x, y))
	newBoxCanvas.Resize(fyne.NewSize(BoxSize, BoxSize))
	return newBoxCanvas
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
	if _, c := GlobalBoard[e]; c {
		return
	}
	defer Container.Refresh()
	defer EdgeButtons[e].Hide()
	nowStep := len(GlobalBoard)
	obtainsBoxes := GlobalBoard.ObtainsBoxes(e)
	score := len(obtainsBoxes)
	for _, box := range obtainsBoxes {
		BoxesCanvases[box].FillColor = FilledColor[NowTurn]
		BoxesFilledColor[box] = FilledColor[NowTurn]
	}
	EdgesCanvases[e].StrokeColor = HighLightColor[NowTurn]
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
					BoxesCanvases[box].FillColor = BoxesFilledColor[box]
					BoxesCanvases[box].Refresh()
				}()
				ticker := time.NewTicker(AnimationStepTime)
				defer ticker.Stop()
				for {
					for i := 0; i <= AnimationSteps; i++ {
						if nowStep != len(GlobalBoard) {
							return
						}
						t := float64(i) / float64(AnimationSteps)
						BoxesCanvases[box].FillColor = interpolateColor(TipColor, getThemeColor(), t)
						BoxesCanvases[box].Refresh()
						<-ticker.C
					}
					for i := 0; i <= AnimationSteps; i++ {
						if nowStep != len(GlobalBoard) {
							return
						}
						t := float64(i) / float64(AnimationSteps)
						BoxesCanvases[box].FillColor = interpolateColor(getThemeColor(), TipColor, t)
						BoxesCanvases[box].Refresh()
						<-ticker.C
					}
				}
			}()
		}
	}
	if nowStep == EdgesCount {
		if PlayerScore[Player1Turn] > PlayerScore[Player2Turn] {
			colog.Info("Player1 Win!")
		} else if PlayerScore[Player1Turn] < PlayerScore[Player2Turn] {
			colog.Info("Player2 Win!")
		} else if PlayerScore[Player1Turn] == PlayerScore[Player2Turn] {
			colog.Infof("Draw!")
		}
		MainWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
			MainWindow.Close()
		})
		go func() {
			time.Sleep(10 * time.Second)
			MainWindow.Close()
		}()
		return
	}
	if AIPlayer1 && NowTurn == Player1Turn {
		SignChan <- struct{}{}
	} else if AIPlayer2 && NowTurn == Player2Turn {
		SignChan <- struct{}{}
	}
}

func GetNextEdges(board Board) (bestEdge Edge) {
	enemyMinScore := 3
	for e := range EdgesSet {
		if _, c := board[e]; !c {
			if score := board.ObtainsScore(e); score > 0 {
				return e
			} else if score == 0 {
				boxes := e.NearBoxes()
				enemyScore := 0
				for _, box := range boxes {
					if board.EdgesCountInBox(box) == 2 {
						enemyScore++
					}
				}
				if enemyMinScore > enemyScore {
					enemyMinScore = enemyScore
					bestEdge = e
				}
			}
		}
	}
	return
}

func GetBestEdge(board Board) (bestEdge Edge) {
	var t atomic.Int64
	var wg sync.WaitGroup
	var latch sync.Mutex
	searchTime := make(map[Edge]int)
	sumScore := make(map[Edge]int)
	wg.Add(Goroutines)
	depth := int64(EdgesCount - len(board))
	for range Goroutines {
		go func() {
			defer wg.Done()
			for t.Load() < SearchTime {
				b := NewBoard(board)
				t.Add(depth)
				turn := Player1Turn
				var firstEdge Edge
				var score int
				for len(b) < EdgesCount {
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
				searchTime[firstEdge]++
				sumScore[firstEdge] += score
				latch.Unlock()
			}
		}()
	}
	wg.Wait()
	bestScore := -1e9
	for e, score := range sumScore {
		averageScore := float64(score) / float64(searchTime[e])
		if averageScore > bestScore {
			bestEdge = e
			bestScore = averageScore
		}
	}
	return
}

type GameTheme struct{}

func (*GameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if GlobalSystemColor != variant {
		GlobalSystemColor = variant
		for _, circle := range DotCanvases {
			circle.FillColor = getDotCanvasColor()
		}
		for box, rectangle := range BoxesCanvases {
			if _, c := BoxesFilledColor[box]; c {
				rectangle.FillColor = getThemeColor()
			}
		}
		Container.Refresh()
	}
	switch name {
	case theme.ColorNameBackground:
		return getThemeColor()
	case theme.ColorNameButton:
		return getButtonColor()
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (*GameTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }

func (*GameTheme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

func (*GameTheme) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

func main() {
	if Record {
		logFilePath := filepath.Join("game log", time.Now().Format(time.DateTime)+".log")
		if err := colog.OpenLog(logFilePath); err != nil {
			panic(err)
		}
	}
	App.Settings().SetTheme(&GameTheme{})
	Container = container.NewWithoutLayout()
	for _, b := range Boxes {
		BoxesCanvases[b] = NewBoxCanvas(b)
		Container.Add(BoxesCanvases[b])
	}
	for e := range EdgesSet {
		EdgesCanvases[e] = NewEdgeCanvas(e)
		Container.Add(EdgesCanvases[e])
		EdgeButtons[e] = widget.NewButton("", func() {
			if AIPlayer1 && NowTurn == Player1Turn {
				return
			} else if AIPlayer2 && NowTurn == Player2Turn {
				return
			}
			AddEdge(e)
		})
		if e.Dot1().X() == e.Dot2().X() {
			EdgeButtons[e].Resize(fyne.NewSize(DotWidth, DotDistance))
			PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - EdgeButtons[e].Size().Width/2 + DotWidth/2
			PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - EdgeButtons[e].Size().Height/2 + DotWidth/2
			EdgeButtons[e].Move(fyne.NewPos(PosX, PosY))
		} else {
			EdgeButtons[e].Resize(fyne.NewSize(DotDistance, DotWidth))
			PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - EdgeButtons[e].Size().Width/2 + DotWidth/2
			PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - EdgeButtons[e].Size().Height/2 + DotWidth/2
			EdgeButtons[e].Move(fyne.NewPos(PosX, PosY))
		}
		Container.Add(EdgeButtons[e])
	}
	for _, d := range Dots {
		DotCanvases[d] = NewDotCanvas(d)
		Container.Add(DotCanvases[d])
	}
	MainWindow.Resize(fyne.NewSize(MainWindowSize, MainWindowSize))
	MainWindow.SetContent(Container)
	MainWindow.SetFixedSize(true)
	go func() {
		if AIPlayer1 {
			AddEdge(GetBestEdge(GlobalBoard))
		}
		for range SignChan {
			if AIPlayer1 && NowTurn == Player1Turn {
				AddEdge(GetBestEdge(GlobalBoard))
			} else if AIPlayer2 && NowTurn == Player2Turn {
				AddEdge(GetBestEdge(GlobalBoard))
			}
		}
	}()
	colog.Info("GAME START!")
	MainWindow.ShowAndRun()
}
