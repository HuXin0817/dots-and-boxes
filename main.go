package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	MaxStepTime       = time.Second
	AnimationSteps    = 100
	AnimationStepTime = time.Second / time.Duration(AnimationSteps)
)

var (
	Goroutines        = runtime.NumCPU()
	BoardSize         = 6
	BoardSizePower    = Dot(BoardSize * BoardSize)
	AIPlayer1         atomic.Bool
	AIPlayer2         atomic.Bool
	DotDistance       = float32(80)
	DotWidth          = DotDistance / 5
	DotMargin         = DotDistance / 3 * 2
	BoxSize           = DotDistance - DotWidth
	MainWindowSize    = DotDistance*float32(BoardSize) + DotMargin - 5
	App               = app.New()
	MainWindow        = App.NewWindow("Dots and Boxes")
	GlobalSystemColor fyne.ThemeVariant
	LockState         = false
	EdgesCount        int
	Dots              []Dot
	Boxes             []Box
	FullBoard         = make(Board)
	EdgeNearBoxes     = make(map[Edge][]Box)
	BoxEdges          = make(map[Box][]Edge)
	DotCanvases       = make(map[Dot]*canvas.Circle)
	EdgesCanvases     = make(map[Edge]*canvas.Line)
	BoxesCanvases     = make(map[Box]*canvas.Rectangle)
	EdgeButtons       = make(map[Edge]*widget.Button)
	BoxesFilledColor  = make(map[Box]color.Color)
	Container         = container.NewWithoutLayout()
	SignChan          = make(chan struct{}, 1)
	MoveRecord        []Edge
	NowTurn           = Player1Turn
	PlayerScore       = map[Turn]int{Player1Turn: 0, Player2Turn: 0}
	GlobalBoard       = make(Board)
	mu                sync.Mutex
	TipColor          = color.NRGBA{R: 255, G: 255, B: 30, A: 50}
	HighLightColor    = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 255, A: 128},
		Player2Turn: {R: 255, G: 30, B: 30, A: 128},
	}
	FilledColor = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 128, A: 128},
		Player2Turn: {R: 128, G: 30, B: 30, A: 128},
	}
)

type Turn int

const (
	Player1Turn Turn = 1
	Player2Turn      = -Player1Turn
)

func (t *Turn) ToString() string {
	if *t == Player1Turn {
		return "Player1"
	} else {
		return "Player2"
	}
}

func (t *Turn) Change() { *t = -*t }

type Dot int

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

func (d Dot) ToString() string { return fmt.Sprintf("(%d, %d)", d.X(), d.Y()) }

type Edge int

func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*BoardSizePower + Dot2) }

func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

func (e Edge) ToString() string { return e.Dot1().ToString() + " => " + e.Dot2().ToString() }

func (e Edge) NearBoxes() []Box { return EdgeNearBoxes[e] }

type Box int

func (b Box) Edges() []Edge { return BoxEdges[b] }

type Board map[Edge]struct{}

func NewBoard(board Board) Board {
	b := make(Board, len(board))
	for e := range board {
		b[e] = struct{}{}
	}
	return b
}

func EdgesCountInBox(b Board, box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b[e]; c {
			count++
		}
	}
	return
}

func ObtainsScore(b Board, e Edge) (count int) {
	boxes := e.NearBoxes()
	for _, box := range boxes {
		if EdgesCountInBox(b, box) == 3 {
			count++
		}
	}
	return
}

func ObtainsBoxes(b Board, e Edge) (obtainsBoxes []Box) {
	boxes := e.NearBoxes()
	for _, box := range boxes {
		if EdgesCountInBox(b, box) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}
	return
}

func SetBoardSize(size int) {
	BoardSize = size
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
	NewGame()
	log.Printf("Game Start! BoardSize: %d\n", BoardSize)
}

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

func getColorByVariant(lightColor, darkColor color.Color) color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return darkColor
	} else {
		return lightColor
	}
}

func GetDotCanvasColor() color.Color {
	return getColorByVariant(color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{R: 202, G: 202, B: 202, A: 255})
}

func GetThemeColor() color.Color {
	return getColorByVariant(color.RGBA{R: 242, G: 242, B: 242, A: 255}, color.RGBA{R: 43, G: 43, B: 43, A: 255})
}

func GetButtonColor() color.Color {
	return getColorByVariant(color.RGBA{R: 217, G: 217, B: 217, A: 255}, color.RGBA{R: 65, G: 65, B: 65, A: 255})
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
	x := transPosition(d.X()) + DotWidth - 0.2
	y := transPosition(d.Y()) + DotWidth - 0.2
	newBoxCanvas := canvas.NewRectangle(GetThemeColor())
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

func GetNextEdges(board Board) (bestEdge Edge) {
	enemyMinScore := 3
	for e := range FullBoard {
		if _, c := board[e]; !c {
			if score := ObtainsScore(board, e); score > 0 {
				return e
			} else if score == 0 {
				boxes := e.NearBoxes()
				enemyScore := 0
				for _, box := range boxes {
					if EdgesCountInBox(board, box) == 2 {
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

func Search(b Board) (firstEdge Edge, score int) {
	turn := Player1Turn
	for len(b) < EdgesCount {
		edge := GetNextEdges(b)
		if firstEdge == 0 {
			firstEdge = edge
		}
		s := ObtainsScore(b, edge)
		score += int(turn) * s
		if s == 0 {
			turn.Change()
		}
		b[edge] = struct{}{}
	}
	return
}

func GetBestEdge() (bestEdge Edge) {
	var (
		globalSearchTime = make(map[Edge]int)
		globalSumScore   = make(map[Edge]int)
		localSearchTimes = make([]map[Edge]int, Goroutines)
		localSumScores   = make([]map[Edge]int, Goroutines)
		wg               sync.WaitGroup
	)
	wg.Add(Goroutines)
	for i := 0; i < Goroutines; i++ {
		localSearchTime := make(map[Edge]int)
		localSumScore := make(map[Edge]int)
		localSearchTimes[i] = localSearchTime
		localSumScores[i] = localSumScore
		go func(localSearchTime, localSumScore map[Edge]int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), MaxStepTime)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					b := NewBoard(GlobalBoard)
					firstEdge, score := Search(b)
					localSearchTime[firstEdge]++
					localSumScore[firstEdge] += score
				}
			}
		}(localSearchTime, localSumScore)
	}
	wg.Wait()
	for i := range Goroutines {
		for e, s := range localSearchTimes[i] {
			globalSearchTime[e] += s
		}
		for e, s := range localSumScores[i] {
			globalSumScore[e] += s
		}
	}
	bestScore := -1e9
	for e, score := range globalSumScore {
		averageScore := float64(score) / float64(globalSearchTime[e])
		if averageScore > bestScore {
			bestEdge = e
			bestScore = averageScore
		}
	}
	return
}

func NewGame() {
	LockState = false
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
			defer Container.Refresh()
			if AIPlayer1.Load() && NowTurn == Player1Turn {
				return
			} else if AIPlayer2.Load() && NowTurn == Player2Turn {
				return
			}
			AddEdge(e, true)
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
		if AIPlayer1.Load() {
			mu.Lock()
			AddEdge(GetBestEdge(), true)
			Container.Refresh()
			mu.Unlock()
		}
		for range SignChan {
			mu.Lock()
			if AIPlayer1.Load() && NowTurn == Player1Turn {
				AddEdge(GetBestEdge(), true)
			} else if AIPlayer2.Load() && NowTurn == Player2Turn {
				AddEdge(GetBestEdge(), true)
			}
			Container.Refresh()
			mu.Unlock()
		}
	}()
	MainWindow.SetContent(Container)
}

func AddEdge(e Edge, withLog bool) {
	if BoardSize <= 1 {
		return
	}
	if _, c := GlobalBoard[e]; c {
		return
	}
	defer func() { MoveRecord = append(MoveRecord, e) }()
	defer EdgeButtons[e].Hide()
	nowStep := len(GlobalBoard)
	obtainsBoxes := ObtainsBoxes(GlobalBoard, e)
	score := len(obtainsBoxes)
	for _, box := range obtainsBoxes {
		BoxesCanvases[box].FillColor = FilledColor[NowTurn]
		BoxesFilledColor[box] = FilledColor[NowTurn]
	}
	EdgesCanvases[e].StrokeColor = HighLightColor[NowTurn]
	PlayerScore[NowTurn] += score
	if withLog {
		log.Printf("Step: %d, Turn %s, Edge: %s, Player1 Score: %d, Player2 Score: %d\n", nowStep, NowTurn.ToString(), e.ToString(), PlayerScore[Player1Turn], PlayerScore[Player2Turn])
	}
	if score == 0 {
		NowTurn.Change()
	}
	GlobalBoard[e] = struct{}{}
	nowStep++
	for _, box := range Boxes {
		edgesCountInBox := EdgesCountInBox(GlobalBoard, box)
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
						BoxesCanvases[box].FillColor = interpolateColor(TipColor, GetThemeColor(), t)
						BoxesCanvases[box].Refresh()
						<-ticker.C
					}
					for i := 0; i <= AnimationSteps; i++ {
						if nowStep != len(GlobalBoard) {
							return
						}
						t := float64(i) / float64(AnimationSteps)
						BoxesCanvases[box].FillColor = interpolateColor(GetThemeColor(), TipColor, t)
						BoxesCanvases[box].Refresh()
						<-ticker.C
					}
				}
			}()
		}
	}
	if nowStep == EdgesCount {
		if PlayerScore[Player1Turn] > PlayerScore[Player2Turn] {
			log.Println("Player1 Win!")
		} else if PlayerScore[Player1Turn] < PlayerScore[Player2Turn] {
			log.Println("Player2 Win!")
		} else if PlayerScore[Player1Turn] == PlayerScore[Player2Turn] {
			log.Println("Draw!")
		}
		return
	}
	if AIPlayer1.Load() && NowTurn == Player1Turn {
		select {
		case SignChan <- struct{}{}:
		default:
		}
	} else if AIPlayer2.Load() && NowTurn == Player2Turn {
		select {
		case SignChan <- struct{}{}:
		default:
		}
	}
}

func StartAIPlayer1() {
	if !AIPlayer1.Load() {
		log.Println("AIPlayer1 ON")
		if NowTurn == Player1Turn {
			select {
			case SignChan <- struct{}{}:
			default:
			}
		}
	} else {
		log.Println("AIPlayer1 OFF")
	}
	AIPlayer1.Store(!AIPlayer1.Load())
}

func StartAIPlayer2() {
	if !AIPlayer2.Load() {
		log.Println("AIPlayer2 ON")
		if NowTurn == Player2Turn {
			select {
			case SignChan <- struct{}{}:
			default:
			}
		}
	} else {
		log.Println("AIPlayer2 OFF")
	}
	AIPlayer2.Store(!AIPlayer2.Load())
}

func Recover(MoveRecord []Edge) {
	NewGame()
	for _, e := range MoveRecord {
		AddEdge(e, false)
	}
}

type GameTheme struct{}

func (GameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if GlobalSystemColor != variant {
		GlobalSystemColor = variant
		for _, circle := range DotCanvases {
			circle.FillColor = GetDotCanvasColor()
		}
		for box, rectangle := range BoxesCanvases {
			if _, c := BoxesFilledColor[box]; !c {
				rectangle.FillColor = GetThemeColor()
			}
		}
		Container.Refresh()
	}
	switch name {
	case theme.ColorNameBackground:
		return GetThemeColor()
	case theme.ColorNameButton:
		return GetButtonColor()
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (GameTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }

func (GameTheme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

func (GameTheme) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

func main() {
	MainWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		if event.Name == fyne.KeySpace {
			if !LockState {
				log.Println("Game Paused")
				mu.Lock()
			} else {
				log.Println("Game Continues")
				mu.Unlock()
			}
			LockState = !LockState
			return
		}
		mu.Lock()
		defer mu.Unlock()
		defer Container.Refresh()
		switch event.Name {
		case fyne.KeyR:
			SetBoardSize(BoardSize)
		case fyne.Key1:
			StartAIPlayer1()
		case fyne.Key2:
			StartAIPlayer2()
		case fyne.KeyUp:
			SetBoardSize(BoardSize + 1)
		case fyne.KeyDown:
			if BoardSize <= 1 {
				return
			}
			SetBoardSize(BoardSize - 1)
		case fyne.KeyLeft:
			if DotDistance-10 <= 0 {
				return
			}
			SetDotDistance(DotDistance - 10)
		case fyne.KeyRight:
			SetDotDistance(DotDistance + 10)
		case fyne.KeyZ:
			moveRecord := append([]Edge{}, MoveRecord...)
			if len(moveRecord) > 0 {
				e := moveRecord[len(moveRecord)-1]
				moveRecord = moveRecord[:len(moveRecord)-1]
				log.Printf("Undo Edge %s\n", e.ToString())
				Recover(moveRecord)
			}
		case fyne.KeyW:
			if BoardSize != 6 {
				SetBoardSize(6)
			}
		case fyne.KeyQ:
			MainWindow.Close()
			log.Println("Game Closed")
		default:
			log.Println("Unidentified Input Key:", event.Name)
		}
	})
	MainWindow.SetFixedSize(true)
	SetBoardSize(6)
	App.Settings().SetTheme(GameTheme{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		Container.Refresh()
	}()
	MainWindow.ShowAndRun()
}
