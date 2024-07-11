package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const HelpDoc = `
Dots and Boxes Game Instructions

Objective:
- The goal of the game is to form more boxes than your opponent.

Gameplay:
- The game is played on a grid of dots.
- Two players take turns connecting two adjacent dots with a horizontal or vertical line.
- If a player completes the fourth side of a box, they score a point and take another turn.
- The game continues until all possible lines are drawn.
- The player with the most completed boxes at the end of the game wins.

Controls:
- Click on the edges between dots to draw a line.
- Use the menu options to control game settings and options.

Menu Options:
- Game:
  - Restart: Start a new game with the current board size.
  - Undo: Undo the last move made.
  - Pause: Pause the game.
  - Score: Display the current score.
  - Quit: Exit the game.

- Board:
  - Add Board Size: Increase the size of the board.
  - Reduce Board Size: Decrease the size of the board.
  - Reset Board: Reset the board to default size and settings.
  
- Config:
  - AI Player 1: Toggle AI for Player 1.
  - AI Player 2: Toggle AI for Player 2.
  - Auto Restart: Automatically restart the game after it ends.
  - Music: Toggle background music on/off.

- Help:
  - Help: Display this help document.

Tips:
- Plan your moves ahead to avoid giving your opponent an opportunity to complete a box.
- Use the edges wisely to maximize your chances of completing boxes.

Have fun playing Dots and Boxes!
`

const (
	AnimationSteps     = 100
	AnimationStepTime  = time.Second / time.Duration(AnimationSteps)
	ImagePath          = "Dots-and-Boxes.png"
	DefaultDotDistance = 80
	DefaultDotSize     = 6
	MinDotSize         = 60
)

var (
	MaxStepTime              = time.Second
	Goroutines               = runtime.NumCPU()
	App                      = app.New()
	MainWindow               = App.NewWindow("Dots and Boxes")
	BoardSize                int
	BoardSizePower           Dot
	DotDistance              float32
	DotWidth                 float32
	DotMargin                float32
	BoxSize                  float32
	MainWindowSize           float32
	GlobalThemeVariant       fyne.ThemeVariant
	EdgesCount               int
	Dots                     []Dot
	Boxes                    []Box
	FullBoard                Board
	EdgeNearBoxes            map[Edge][]Box
	BoxEdges                 map[Box][]Edge
	DotCanvases              map[Dot]*canvas.Circle
	EdgesCanvases            map[Edge]*canvas.Line
	BoxesCanvases            map[Box]*canvas.Rectangle
	EdgeButtons              map[Edge]*widget.Button
	BoxesFilledColor         map[Box]color.Color
	PlayerScore              map[Turn]int
	Container                *fyne.Container
	SignChan                 chan struct{}
	MoveRecord               []Edge
	GlobalBoard              Board
	NowTurn                  Turn
	mu                       sync.Mutex
	boxesCanvasMu            sync.Mutex
	AIPlayer1                = NewOption("AIPlayer1", false)
	AIPlayer2                = NewOption("AIPlayer2", false)
	AutoRestart              = NewOption("AutoRestart", false)
	MusicOn                  = NewOption("Music", true)
	LightThemeDotCanvasColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	DarkThemeDotCanvasColor  = color.NRGBA{R: 202, G: 202, B: 202, A: 255}
	LightThemeColor          = color.NRGBA{R: 242, G: 242, B: 242, A: 255}
	DarkThemeColor           = color.NRGBA{R: 43, G: 43, B: 43, A: 255}
	LightThemeButtonColor    = color.NRGBA{R: 217, G: 217, B: 217, A: 255}
	DarkThemeButtonColor     = color.NRGBA{R: 65, G: 65, B: 65, A: 255}
	Player1HighlightColor    = color.NRGBA{R: 64, G: 64, B: 255, A: 128}
	Player2HighlightColor    = color.NRGBA{R: 255, G: 64, B: 64, A: 128}
	Player1FilledColor       = color.NRGBA{R: 64, G: 64, B: 255, A: 64}
	Player2FilledColor       = color.NRGBA{R: 255, G: 64, B: 64, A: 64}
	TipColor                 = color.NRGBA{R: 255, G: 255, B: 64, A: 64}
	RestartMenuItem          *fyne.MenuItem
	MusicMenuItem            *fyne.MenuItem
	AIPlayer1MenuItem        *fyne.MenuItem
	AIPlayer2MenuItem        *fyne.MenuItem
	AutoRestartMenuItem      *fyne.MenuItem
	PauseMenuItem            *fyne.MenuItem
	AddBoardSizeMenuItem     *fyne.MenuItem
	ReduceBoardSizeMenuItem  *fyne.MenuItem
	UndoMenuItem             *fyne.MenuItem
	AddBoardWidthMenuItem    *fyne.MenuItem
	ReduceBoardWidthMenuItem *fyne.MenuItem
	ResetBoardMenuItem       *fyne.MenuItem
	ScoreMenuItem            *fyne.MenuItem
	QuitMenuItem             *fyne.MenuItem
	HelpMenuItem             *fyne.MenuItem
	PauseState               bool
	TmpMenuItemDisabledState = make(map[string]bool)
	UnDisableMenuItemLabel   = map[string]struct{}{
		"Continue":  {},
		"Quit":      {},
		"Help":      {},
		"Music ON":  {},
		"Music OFF": {},
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

func SendMessage(format string, a ...any) {
	App.SendNotification(&fyne.Notification{
		Title:   "Dots-And-Boxes",
		Content: fmt.Sprintf(format, a...),
	})
}

type Option struct {
	name  string
	value atomic.Bool
}

func NewOption(name string, value bool) *Option {
	op := &Option{name: name}
	op.value.Store(value)
	return op
}

func (op *Option) Value() bool { return op.value.Load() }

func (op *Option) On() {
	if !op.value.Load() {
		op.Change()
	}
}

func (op *Option) Off() {
	if op.value.Load() {
		op.Change()
	}
}

func (op *Option) Change() {
	if op.value.Load() {
		SendMessage(op.name + " OFF")
	} else {
		SendMessage(op.name + " ON")
	}
	op.value.Store(!op.value.Load())
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
	globalSearchTime := make(map[Edge]int)
	globalSumScore := make(map[Edge]int)
	localSearchTimes := make([]map[Edge]int, Goroutines)
	localSumScores := make([]map[Edge]int, Goroutines)
	var wg sync.WaitGroup
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

type Music struct {
	*bytes.Reader
}

func (rc *Music) Close() error { return nil }

func moveMusic() *Music { return &Music{bytes.NewReader(moveMp3)} }

func scoreMusic() *Music { return &Music{bytes.NewReader(scoreMp3)} }

func PlayMoveMusic() { play(moveMusic()) }

func PlayScoreMusic() { play(scoreMusic()) }

var musicLock sync.Mutex

func play(m *Music) {
	if MusicOn.Value() {
		musicLock.Lock()
		defer musicLock.Unlock()
		streamer, format, err := mp3.Decode(m)
		if err != nil {
			fmt.Println("Error decoding file:", err)
			return
		}
		defer streamer.Close()
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() { done <- true })))
		<-done
	}
}

type GameTheme struct{}

func (GameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if GlobalThemeVariant != variant {
		GlobalThemeVariant = variant
		for _, circle := range DotCanvases {
			circle.FillColor = GetDotCanvasColor()
			circle.Refresh()
		}
		for box, rectangle := range BoxesCanvases {
			if _, c := BoxesFilledColor[box]; !c {
				rectangle.FillColor = GetThemeColor()
				rectangle.Refresh()
			}
		}
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

func (GameTheme) Icon(fyne.ThemeIconName) fyne.Resource { return MainWindow.Icon() }

func (GameTheme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

func (GameTheme) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	r := uint8((float64(r1)*(1-t) + float64(r2)*t) / 256)
	g := uint8((float64(g1)*(1-t) + float64(g2)*t) / 256)
	b := uint8((float64(b1)*(1-t) + float64(b2)*t) / 256)
	a := uint8((float64(a1)*(1-t) + float64(a2)*t) / 256)
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func getColorByVariant(lightColor, darkColor color.Color) color.Color {
	if GlobalThemeVariant == theme.VariantDark {
		return darkColor
	} else {
		return lightColor
	}
}

func GetDotCanvasColor() color.Color {
	return getColorByVariant(LightThemeDotCanvasColor, DarkThemeDotCanvasColor)
}

func GetThemeColor() color.Color {
	return getColorByVariant(LightThemeColor, DarkThemeColor)
}

func GetButtonColor() color.Color {
	return getColorByVariant(LightThemeButtonColor, DarkThemeButtonColor)
}

func GetPlayerFilledColor() color.Color {
	if NowTurn == Player1Turn {
		return Player1FilledColor
	} else {
		return Player2FilledColor
	}
}

func GetPlayerHighlightColor() color.Color {
	if NowTurn == Player1Turn {
		return Player1HighlightColor
	} else {
		return Player2HighlightColor
	}
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
	if _, c := GlobalBoard[e]; c {
		return
	}
	if e == 0 {
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

func main() {
	RestartMenuItem = &fyne.MenuItem{
		Label: "Restart",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			defer Refresh()
			RestartWithCall(BoardSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	PauseMenuItem = &fyne.MenuItem{
		Label:    "Pause",
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeySpace},
		Action: func() {
			defer MainWindow.MainMenu().Refresh()
			if !PauseState {
				mu.Lock()
				PauseMenuItem.Label = "Continue"
				SendMessage("Game Paused")
				TmpMenuItemDisabledState = make(map[string]bool)
				for _, menu := range MainWindow.MainMenu().Items {
					for _, m := range menu.Items {
						if _, c := UnDisableMenuItemLabel[m.Label]; !c {
							TmpMenuItemDisabledState[m.Label] = m.Disabled
							m.Disabled = true
						}
					}
				}
			} else {
				for _, menu := range MainWindow.MainMenu().Items {
					for _, m := range menu.Items {
						if _, c := UnDisableMenuItemLabel[m.Label]; !c {
							m.Disabled = TmpMenuItemDisabledState[m.Label]
						}
					}
				}
				mu.Unlock()
				PauseMenuItem.Label = "Pause"
				SendMessage("Game Continue")
			}
			PauseState = !PauseState
		},
	}

	ScoreMenuItem = &fyne.MenuItem{
		Label: "Score",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			SendMessage("Player1 Score: %d\nPlayer2 Score: %d\n", PlayerScore[Player1Turn], PlayerScore[Player2Turn])
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS},
	}

	AddBoardSizeMenuItem = &fyne.MenuItem{
		Label: "AddBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			RestartWithCall(BoardSize + 1)
			if !ReduceBoardSizeMenuItem.Disabled {
				ReduceBoardSizeMenuItem.Disabled = false
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyEqual},
	}

	ReduceBoardSizeMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			defer Refresh()
			if BoardSize <= 1 {
				return
			}
			RestartWithCall(BoardSize - 1)
			if BoardSize <= 1 {
				ResetBoardMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyMinus},
	}

	UndoMenuItem = &fyne.MenuItem{
		Label: "Undo",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			moveRecord := append([]Edge{}, MoveRecord...)
			if len(moveRecord) > 0 {
				e := moveRecord[len(moveRecord)-1]
				SendMessage("Undo Edge " + e.ToString())
				moveRecord = moveRecord[:len(moveRecord)-1]
				Recover(moveRecord)
			}
			if len(moveRecord) == 0 {
				UndoMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Disabled: true,
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyZ},
	}

	AddBoardWidthMenuItem = &fyne.MenuItem{
		Label: "AddBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			SetDotDistance(DotDistance + 10)
			if ReduceBoardWidthMenuItem.Disabled {
				ReduceBoardWidthMenuItem.Disabled = true
			}
			if BoardSize != DefaultDotSize || DotDistance != DefaultDotSize {
				ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyUp},
	}

	ReduceBoardWidthMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			if DotDistance-10 < MinDotSize {
				return
			}
			SetDotDistance(DotDistance - 10)
			if DotDistance-10 < MinDotSize {
				ReduceBoardSizeMenuItem.Disabled = true
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyDown},
	}

	ResetBoardMenuItem = &fyne.MenuItem{
		Label: "ResetBoard",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			defer MainWindow.MainMenu().Refresh()
			ResetBoard()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	AIPlayer1MenuItem = &fyne.MenuItem{
		Label: "AIPlayer1 ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AIPlayer1.Value() {
				defer notifySignChan()
				AIPlayer1MenuItem.Label = "AIPlayer1 OFF"
			} else {
				AIPlayer1MenuItem.Label = "AIPlayer1 ON"
			}
			AIPlayer1.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key1},
	}

	AIPlayer2MenuItem = &fyne.MenuItem{
		Label: "AIPlayer2 ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AIPlayer2.Value() {
				defer notifySignChan()
				AIPlayer2MenuItem.Label = "AIPlayer2 OFF"
			} else {
				AIPlayer2MenuItem.Label = "AIPlayer2 ON"
			}
			AIPlayer2.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key2},
	}

	MusicMenuItem = &fyne.MenuItem{
		Label: "Music OFF",
		Action: func() {
			defer MainWindow.MainMenu().Refresh()
			if !MusicOn.Value() {
				MusicMenuItem.Label = "Music OFF"
			} else {
				MusicMenuItem.Label = "Music ON"
			}
			MusicOn.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyP},
	}

	AutoRestartMenuItem = &fyne.MenuItem{
		Label: "AutoRestart ON",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer MainWindow.MainMenu().Refresh()
			if !AutoRestart.Value() {
				AutoRestartMenuItem.Label = "AutoRestart OFF"
			} else {
				AutoRestartMenuItem.Label = "AutoRestart ON"
			}
			AutoRestart.Change()
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyA},
	}

	QuitMenuItem = &fyne.MenuItem{
		Label: "Quit",
		Action: func() {
			SendMessage("Game Closed")
			os.Exit(0)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyQ},
		IsQuit:   true,
	}

	HelpMenuItem = &fyne.MenuItem{
		Label: "Help",
		Action: func() {
			SendMessage(HelpDoc)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyH},
	}

	MainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Game",
			RestartMenuItem,
			UndoMenuItem,
			PauseMenuItem,
			ScoreMenuItem,
			fyne.NewMenuItemSeparator(),
			QuitMenuItem,
		),
		fyne.NewMenu("Board",
			AddBoardWidthMenuItem,
			ReduceBoardWidthMenuItem,
			AddBoardSizeMenuItem,
			ReduceBoardSizeMenuItem,
			fyne.NewMenuItemSeparator(),
			ResetBoardMenuItem,
		),
		fyne.NewMenu("Config",
			AIPlayer1MenuItem,
			AIPlayer2MenuItem,
			AutoRestartMenuItem,
			MusicMenuItem,
		),
		fyne.NewMenu("Help",
			HelpMenuItem,
		),
	))

	ResetBoard()
	MainWindow.SetFixedSize(true)
	App.Settings().SetTheme(GameTheme{})
	go func() {
		time.Sleep(300 * time.Millisecond)
		Container.Refresh()
	}()
	MainWindow.ShowAndRun()
}
