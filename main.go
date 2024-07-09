package main

import (
	"context"
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

const (
	MaxStepTime       = 1 * time.Second
	Goroutines        = 32
	AnimationSteps    = 100
	AnimationStepTime = time.Second / time.Duration(AnimationSteps)
	DotDistance       = float32(80)
	DotWidth          = DotDistance / 5
	DotMargin         = DotDistance / 3 * 2
	BoxSize           = DotDistance - DotWidth
)

var (
	BoardSize         = 6
	BoardSizePower    = Dot(BoardSize * BoardSize)
	AIPlayer1         atomic.Bool
	AIPlayer2         atomic.Bool
	MainWindowSize    = DotDistance*float32(BoardSize) + DotMargin - 5
	App               = app.New()
	MainWindow        = App.NewWindow("Dots and Boxes")
	NowGame           = &Game{}
	Dots              []Dot
	EdgesCount        int
	Edges             []Edge
	EdgesSet          map[Edge]struct{}
	Boxes             []Box
	EdgeNearBoxes     map[Edge][]Box
	BoxEdges          map[Box][]Edge
	GlobalSystemColor fyne.ThemeVariant
	LogRecord         bool
	HighLightColor    = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 255, A: 128},
		Player2Turn: {R: 255, G: 30, B: 30, A: 128},
	}
	FilledColor = map[Turn]color.NRGBA{
		Player1Turn: {R: 30, G: 30, B: 128, A: 128},
		Player2Turn: {R: 128, G: 30, B: 30, A: 128},
	}
	TipColor = color.NRGBA{R: 255, G: 255, B: 30, A: 50}
	mu       sync.Mutex
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

func SetBoardSize(size int) {
	mu.Lock()
	defer mu.Unlock()
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
	Edges = []Edge{}
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
	EdgesSet = make(map[Edge]struct{})
	for _, e := range Edges {
		EdgesSet[e] = struct{}{}
	}
	Boxes = []Box{}
	for _, d := range Dots {
		if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
			Boxes = append(Boxes, Box(d))
		}
	}
	EdgeNearBoxes = make(map[Edge][]Box)
	for e := range EdgesSet {
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
	EdgesCount = len(Edges)
}

func transPosition(x int) float32 { return DotMargin + float32(x)*DotDistance }

func GetDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

func GetDotCanvasColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 202, G: 202, B: 202, A: 255}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func GetThemeColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 43, G: 43, B: 43, A: 255}
	}
	return color.RGBA{R: 242, G: 242, B: 242, A: 255}
}

func GetButtonColor() color.Color {
	if GlobalSystemColor == theme.VariantDark {
		return color.RGBA{R: 65, G: 65, B: 65, A: 255}
	}
	return color.RGBA{R: 217, G: 217, B: 217, A: 255}
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
	x := transPosition(d.X()) + DotWidth - 0.5
	y := transPosition(d.Y()) + DotWidth - 0.5
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

func Search(b Board) (firstEdge Edge, score int) {
	turn := Player1Turn
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
	return
}

func GetBestEdge(board Board) (bestEdge Edge) {
	var (
		globalSearchTime = make(map[Edge]int)
		globalSumScore   = make(map[Edge]int)
		localSearchTimes = make([]map[Edge]int, Goroutines)
		localSumScores   = make([]map[Edge]int, Goroutines)
		wg               sync.WaitGroup
	)
	wg.Add(Goroutines)
	for i := range Goroutines {
		localSearchTime := make(map[Edge]int)
		localSumScore := make(map[Edge]int)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), MaxStepTime)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					localSearchTimes[i] = localSearchTime
					localSumScores[i] = localSumScore
					return
				default:
					b := NewBoard(board)
					firstEdge, score := Search(b)
					localSearchTime[firstEdge]++
					localSumScore[firstEdge] += score
				}
			}
		}()
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

type Game struct {
	Finished         bool
	LockState        bool
	DotCanvases      map[Dot]*canvas.Circle
	EdgesCanvases    map[Edge]*canvas.Line
	BoxesCanvases    map[Box]*canvas.Rectangle
	EdgeButtons      map[Edge]*widget.Button
	BoxesFilledColor map[Box]color.Color
	Container        *fyne.Container
	SignChan         chan struct{}
	NowTurn          Turn
	PlayerScore      map[Turn]int
	GlobalBoard      Board
}

func NewGame() *Game {
	game := &Game{
		DotCanvases:      make(map[Dot]*canvas.Circle),
		EdgesCanvases:    make(map[Edge]*canvas.Line),
		BoxesCanvases:    make(map[Box]*canvas.Rectangle),
		EdgeButtons:      make(map[Edge]*widget.Button),
		BoxesFilledColor: make(map[Box]color.Color),
		Container:        container.NewWithoutLayout(),
		SignChan:         make(chan struct{}, 1),
		NowTurn:          Player1Turn,
		PlayerScore:      map[Turn]int{Player1Turn: 0, Player2Turn: 0},
		GlobalBoard:      make(Board),
	}
	for _, b := range Boxes {
		game.BoxesCanvases[b] = NewBoxCanvas(b)
		game.Container.Add(game.BoxesCanvases[b])
	}
	for e := range EdgesSet {
		game.EdgesCanvases[e] = NewEdgeCanvas(e)
		game.Container.Add(game.EdgesCanvases[e])
		game.EdgeButtons[e] = widget.NewButton("", func() {
			if AIPlayer1.Load() && game.NowTurn == Player1Turn {
				return
			} else if AIPlayer2.Load() && game.NowTurn == Player2Turn {
				return
			}
			mu.Lock()
			game.AddEdge(e)
			mu.Unlock()
		})
		if e.Dot1().X() == e.Dot2().X() {
			game.EdgeButtons[e].Resize(fyne.NewSize(DotWidth, DotDistance))
			PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - game.EdgeButtons[e].Size().Width/2 + DotWidth/2
			PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - game.EdgeButtons[e].Size().Height/2 + DotWidth/2
			game.EdgeButtons[e].Move(fyne.NewPos(PosX, PosY))
		} else {
			game.EdgeButtons[e].Resize(fyne.NewSize(DotDistance, DotWidth))
			PosX := (transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2 - game.EdgeButtons[e].Size().Width/2 + DotWidth/2
			PosY := (transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2 - game.EdgeButtons[e].Size().Height/2 + DotWidth/2
			game.EdgeButtons[e].Move(fyne.NewPos(PosX, PosY))
		}
		game.Container.Add(game.EdgeButtons[e])
	}
	for _, d := range Dots {
		game.DotCanvases[d] = NewDotCanvas(d)
		game.Container.Add(game.DotCanvases[d])
	}
	go func() {
		if AIPlayer1.Load() {
			mu.Lock()
			game.AddEdge(GetBestEdge(game.GlobalBoard))
			mu.Unlock()
		}
		for range game.SignChan {
			mu.Lock()
			if AIPlayer1.Load() && game.NowTurn == Player1Turn {
				game.AddEdge(GetBestEdge(game.GlobalBoard))
			} else if AIPlayer2.Load() && game.NowTurn == Player2Turn {
				game.AddEdge(GetBestEdge(game.GlobalBoard))
			}
			mu.Unlock()
		}
	}()
	if LogRecord {
		logFilePath := filepath.Join("game log", time.Now().Format(time.DateTime)+".log")
		if err := colog.OpenLog(logFilePath); err != nil {
			colog.Error(err)
		}
	}
	colog.Info("GAME START!")
	return game
}

func (game *Game) AddEdge(e Edge) {
	if game.Finished {
		return
	}
	if BoardSize <= 1 {
		return
	}
	if _, c := game.GlobalBoard[e]; c {
		return
	}
	defer game.Container.Refresh()
	defer game.EdgeButtons[e].Hide()
	nowStep := len(game.GlobalBoard)
	obtainsBoxes := game.GlobalBoard.ObtainsBoxes(e)
	score := len(obtainsBoxes)
	for _, box := range obtainsBoxes {
		game.BoxesCanvases[box].FillColor = FilledColor[game.NowTurn]
		game.BoxesFilledColor[box] = FilledColor[game.NowTurn]
	}
	game.EdgesCanvases[e].StrokeColor = HighLightColor[game.NowTurn]
	game.PlayerScore[game.NowTurn] += score
	colog.Infof("Step: %d, Turn %s, Edge: %s, Player1 Score: %d, Player2 Score: %d", nowStep, game.NowTurn.ToString(), e.ToString(), game.PlayerScore[Player1Turn], game.PlayerScore[Player2Turn])
	if score == 0 {
		game.NowTurn.Change()
	}
	game.GlobalBoard[e] = struct{}{}
	nowStep++
	for _, box := range Boxes {
		edgesCountInBox := game.GlobalBoard.EdgesCountInBox(box)
		if edgesCountInBox == 3 {
			go func() {
				defer func() {
					game.BoxesCanvases[box].FillColor = game.BoxesFilledColor[box]
					game.BoxesCanvases[box].Refresh()
				}()
				ticker := time.NewTicker(AnimationStepTime)
				defer ticker.Stop()
				for {
					for i := 0; i <= AnimationSteps; i++ {
						if nowStep != len(game.GlobalBoard) {
							return
						}
						t := float64(i) / float64(AnimationSteps)
						game.BoxesCanvases[box].FillColor = interpolateColor(TipColor, GetThemeColor(), t)
						game.BoxesCanvases[box].Refresh()
						<-ticker.C
					}
					for i := 0; i <= AnimationSteps; i++ {
						if nowStep != len(game.GlobalBoard) {
							return
						}
						t := float64(i) / float64(AnimationSteps)
						game.BoxesCanvases[box].FillColor = interpolateColor(GetThemeColor(), TipColor, t)
						game.BoxesCanvases[box].Refresh()
						<-ticker.C
					}
				}
			}()
		}
	}
	if nowStep == EdgesCount {
		if game.PlayerScore[Player1Turn] > game.PlayerScore[Player2Turn] {
			colog.Info("Player1 Win!")
		} else if game.PlayerScore[Player1Turn] < game.PlayerScore[Player2Turn] {
			colog.Info("Player2 Win!")
		} else if game.PlayerScore[Player1Turn] == game.PlayerScore[Player2Turn] {
			colog.Infof("Draw!")
		}
		return
	}
	if AIPlayer1.Load() && game.NowTurn == Player1Turn {
		select {
		case game.SignChan <- struct{}{}:
		default:
		}
	} else if AIPlayer2.Load() && game.NowTurn == Player2Turn {
		select {
		case game.SignChan <- struct{}{}:
		default:
		}
	}
}

func (game *Game) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if GlobalSystemColor != variant {
		GlobalSystemColor = variant
		for _, circle := range game.DotCanvases {
			circle.FillColor = GetDotCanvasColor()
		}
		for box, rectangle := range game.BoxesCanvases {
			if _, c := game.BoxesFilledColor[box]; !c {
				rectangle.FillColor = GetThemeColor()
			}
		}
		game.Container.Refresh()
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

func (*Game) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }

func (*Game) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

func (*Game) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	NowGame.Finished = true
	NowGame = NewGame()
	MainWindow.SetContent(NowGame.Container)
	App.Settings().SetTheme(NowGame)
}

func main() {
	SetBoardSize(6)
	MainWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyR:
			Reset()
		case fyne.Key1:
			mu.Lock()
			if !AIPlayer1.Load() {
				colog.Info("AIPlayer1 ON")
				if NowGame.NowTurn == Player1Turn {
					select {
					case NowGame.SignChan <- struct{}{}:
					default:
					}
				}
			} else {
				colog.Info("AIPlayer1 OFF")
			}
			AIPlayer1.Store(!AIPlayer1.Load())
			mu.Unlock()
		case fyne.Key2:
			mu.Lock()
			if !AIPlayer2.Load() {
				colog.Info("AIPlayer2 ON")
				if NowGame.NowTurn == Player2Turn {
					select {
					case NowGame.SignChan <- struct{}{}:
					default:
					}
				}
			} else {
				colog.Info("AIPlayer2 OFF")
			}
			AIPlayer2.Store(!AIPlayer2.Load())
			mu.Unlock()
		case fyne.KeyUp:
			fallthrough
		case fyne.KeyEqual:
			fallthrough
		case fyne.KeyPlus:
			SetBoardSize(BoardSize + 1)
			Reset()
		case fyne.KeyDown:
			fallthrough
		case fyne.KeyMinus:
			if BoardSize <= 1 {
				return
			}
			SetBoardSize(BoardSize - 1)
			Reset()
		case fyne.KeyW:
			SetBoardSize(6)
			Reset()
		case fyne.KeyQ:
			MainWindow.Close()
			colog.Info("Game Closed")
		case fyne.KeySpace:
			if !NowGame.LockState {
				colog.Info("Game Paused")
				mu.Lock()
			} else {
				colog.Info("Game Continues")
				mu.Unlock()
			}
			NowGame.LockState = !NowGame.LockState
		case fyne.KeyL:
			if LogRecord {
				_ = colog.OpenLog("")
				colog.Info("Log Closed")
			} else {
				colog.Info("Start Log")
			}
			LogRecord = !LogRecord
		default:
			colog.Error("Unidentified Input Key:", event.Name)
		}
	})
	MainWindow.SetFixedSize(true)
	NowGame = NewGame()
	App.Settings().SetTheme(NowGame)
	MainWindow.SetContent(NowGame.Container)
	go func() {
		time.Sleep(100 * time.Millisecond)
		NowGame.Container.Refresh()
	}()
	MainWindow.ShowAndRun()
}
