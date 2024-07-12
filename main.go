package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytedance/sonic"
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
	ConfigFilePath     = "config.json"
	DefaultDotDistance = 80
	DefaultBoardSize   = 6
	DefaultStepTime    = time.Second
	MinDotSize         = 60
	LogRecordFileDir   = "log"
)

type Chess struct {
	BoardSize                int
	BoardSizePower           Dot
	DotWidth                 float32
	DotMargin                float32
	BoxSize                  float32
	MainWindowSize           float32
	DotDistance              float32
	AIPlayer1                bool
	AIPlayer2                bool
	AutoRestart              bool
	MusicOn                  bool
	PauseState               bool
	EdgesCount               int
	MoveRecords              []MoveRecord
	Dots                     []Dot
	Boxes                    []Box
	FullBoard                Board
	EdgeNearBoxes            map[Edge][]Box
	BoxEdges                 map[Box][]Edge
	GlobalBoard              Board
	NowTurn                  Turn
	AISearchTime             time.Duration
	DotCanvases              map[Dot]*canvas.Circle    `json:"-"`
	EdgesCanvases            map[Edge]*canvas.Line     `json:"-"`
	BoxesCanvases            map[Box]*canvas.Rectangle `json:"-"`
	EdgeButtons              map[Edge]*widget.Button   `json:"-"`
	BoxesFilledColor         map[Box]color.Color       `json:"-"`
	PlayerScore              map[Turn]int              `json:"-"`
	Container                *fyne.Container           `json:"-"`
	SignChan                 chan struct{}             `json:"-"`
	RestartMenuItem          *fyne.MenuItem            `json:"-"`
	MusicMenuItem            *fyne.MenuItem            `json:"-"`
	AIPlayer1MenuItem        *fyne.MenuItem            `json:"-"`
	AIPlayer2MenuItem        *fyne.MenuItem            `json:"-"`
	AutoRestartMenuItem      *fyne.MenuItem            `json:"-"`
	PauseMenuItem            *fyne.MenuItem            `json:"-"`
	AddBoardSizeMenuItem     *fyne.MenuItem            `json:"-"`
	ReduceBoardSizeMenuItem  *fyne.MenuItem            `json:"-"`
	UndoMenuItem             *fyne.MenuItem            `json:"-"`
	AddBoardWidthMenuItem    *fyne.MenuItem            `json:"-"`
	ReduceBoardWidthMenuItem *fyne.MenuItem            `json:"-"`
	ResetBoardMenuItem       *fyne.MenuItem            `json:"-"`
	ScoreMenuItem            *fyne.MenuItem            `json:"-"`
	QuitMenuItem             *fyne.MenuItem            `json:"-"`
	HelpMenuItem             *fyne.MenuItem            `json:"-"`
	GlobalThemeVariant       fyne.ThemeVariant         `json:"-"`
}

var chess = func() (chess Chess) {
	if _, err := os.Stat(ConfigFilePath); err == nil {
		if b, err := os.ReadFile(ConfigFilePath); err == nil {
			if err := sonic.Unmarshal(b, &chess); err == nil {
				return chess
			}
		}
	}
	return Chess{
		BoardSize:    DefaultBoardSize,
		DotDistance:  DefaultDotDistance,
		MusicOn:      true,
		AISearchTime: DefaultStepTime,
	}
}()

var (
	Goroutines               = runtime.NumCPU()
	App                      = app.New()
	MainWindow               = App.NewWindow("Dots and Boxes")
	TmpMenuItemDisabledState = make(map[string]bool)
	UnDisableMenuItemLabel   = map[string]struct{}{
		"Continue":  {},
		"Quit":      {},
		"Help":      {},
		"Music ON":  {},
		"Music OFF": {},
	}
)

var (
	mu              sync.Mutex
	boxesCanvasLock sync.Mutex
	musicLock       sync.Mutex

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

func NewDot(x, y int) Dot { return Dot(x*chess.BoardSize + y) }

func (d Dot) X() int { return int(d) / chess.BoardSize }

func (d Dot) Y() int { return int(d) % chess.BoardSize }

func (d Dot) ToString() string { return fmt.Sprintf("(%d, %d)", d.X(), d.Y()) }

type Edge int

func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*chess.BoardSizePower + Dot2) }

func (e Edge) Dot1() Dot { return Dot(e) / chess.BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % chess.BoardSizePower }

func (e Edge) ToString() string { return e.Dot1().ToString() + " => " + e.Dot2().ToString() }

func (e Edge) NearBoxes() []Box { return chess.EdgeNearBoxes[e] }

type Box int

func (b Box) Edges() []Edge { return chess.BoxEdges[b] }

type Board map[Edge]struct{}

func NewBoard(board Board) Board {
	b := make(Board, len(board))
	for e := range board {
		b[e] = struct{}{}
	}
	return b
}

type MoveRecord struct {
	TimeStamp    string
	Step         int
	Player       string
	MoveEdge     Edge
	Player1Score int
	Player2Score int
}

func SendMessage(format string, a ...any) {
	App.SendNotification(&fyne.Notification{
		Title:   "Dots-And-Boxes",
		Content: fmt.Sprintf(format, a...),
	})
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
	for e := range chess.FullBoard {
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
	for len(b) < chess.EdgesCount {
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
			ctx, cancel := context.WithTimeout(context.Background(), chess.AISearchTime)
			defer cancel()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					b := NewBoard(chess.GlobalBoard)
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

func PlayMoveMusic() error { return play(moveMusic()) }

func PlayScoreMusic() error { return play(scoreMusic()) }

func play(m *Music) (err error) {
	if chess.MusicOn {
		musicLock.Lock()
		defer musicLock.Unlock()
		streamer, format, err := mp3.Decode(m)
		if err != nil {
			return err
		}
		defer func(streamer beep.StreamSeekCloser) {
			if err := streamer.Close(); err != nil {
				log.Println(err)
			}
		}(streamer)
		if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
			return err
		}
		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))
		<-done
	}
	return nil
}

type GameTheme struct{}

func (GameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if chess.GlobalThemeVariant != variant {
		chess.GlobalThemeVariant = variant
		for _, circle := range chess.DotCanvases {
			circle.FillColor = GetDotCanvasColor()
			circle.Refresh()
		}
		for box, rectangle := range chess.BoxesCanvases {
			if _, c := chess.BoxesFilledColor[box]; !c {
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
	if chess.GlobalThemeVariant == theme.VariantDark {
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
	if chess.NowTurn == Player1Turn {
		return Player1FilledColor
	} else {
		return Player2FilledColor
	}
}

func GetPlayerHighlightColor() color.Color {
	if chess.NowTurn == Player1Turn {
		return Player1HighlightColor
	} else {
		return Player2HighlightColor
	}
}

func SetDotDistance(d float32) {
	chess.DotDistance = d
	chess.DotWidth = chess.DotDistance / 5
	chess.DotMargin = chess.DotDistance / 3 * 2
	chess.BoxSize = chess.DotDistance - chess.DotWidth
	chess.MainWindowSize = chess.DotDistance*float32(chess.BoardSize) + chess.DotMargin - 5
	MainWindow.Resize(fyne.NewSize(chess.MainWindowSize, chess.MainWindowSize))
	moveRecord := append([]MoveRecord{}, chess.MoveRecords...)
	Recover(moveRecord)
}

func transPosition(x int) float32 { return chess.DotMargin + float32(x)*chess.DotDistance }

func GetDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

func getEdgeButtonSizeAndPosition(e Edge) (size fyne.Size, pos fyne.Position) {
	if e.Dot1().X() == e.Dot2().X() {
		size = fyne.NewSize(chess.DotWidth, chess.DotDistance)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+chess.DotWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+chess.DotWidth/2,
		)
	} else {
		size = fyne.NewSize(chess.DotDistance, chess.DotWidth)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+chess.DotWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+chess.DotWidth/2,
		)
	}
	return
}

func NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(GetDotCanvasColor())
	newDotCanvas.Resize(fyne.NewSize(chess.DotWidth, chess.DotWidth))
	newDotCanvas.Move(fyne.NewPos(GetDotPosition(d)))
	return newDotCanvas
}

func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + chess.DotWidth/2
	y1 := transPosition(e.Dot1().Y()) + chess.DotWidth/2
	x2 := transPosition(e.Dot2().X()) + chess.DotWidth/2
	y2 := transPosition(e.Dot2().Y()) + chess.DotWidth/2
	newEdgeCanvas := canvas.NewLine(GetDotCanvasColor())
	newEdgeCanvas.Position1 = fyne.NewPos(x1, y1)
	newEdgeCanvas.Position2 = fyne.NewPos(x2, y2)
	newEdgeCanvas.StrokeWidth = chess.DotWidth
	return newEdgeCanvas
}

func NewBoxCanvas(box Box) *canvas.Rectangle {
	d := Dot(box)
	x := transPosition(d.X()) + chess.DotWidth
	y := transPosition(d.Y()) + chess.DotWidth
	newBoxCanvas := canvas.NewRectangle(GetThemeColor())
	newBoxCanvas.Move(fyne.NewPos(x, y))
	newBoxCanvas.Resize(fyne.NewSize(chess.BoxSize, chess.BoxSize))
	return newBoxCanvas
}

func Refresh() error {
	defer chess.Container.Refresh()
	defer MainWindow.MainMenu().Refresh()
	img := MainWindow.Canvas().Capture()
	file, err := os.Create(ImagePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return err
	}
	MainWindow.SetIcon(fyne.NewStaticResource("Dots-and-Boxes", buf.Bytes()))
	j, err := sonic.Marshal(chess)
	if err != nil {
		return err
	}
	if err := os.WriteFile(ConfigFilePath, j, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func notifySignChan() {
	if chess.AIPlayer1 && chess.NowTurn == Player1Turn || chess.AIPlayer2 && chess.NowTurn == Player2Turn {
		select {
		case chess.SignChan <- struct{}{}:
		default:
		}
	}
}

func NewGame(boardSize int) {
	chess.BoardSize = boardSize
	chess.BoardSizePower = Dot(chess.BoardSize * chess.BoardSize)
	chess.MainWindowSize = chess.DotDistance*float32(chess.BoardSize) + chess.DotMargin - 5
	MainWindow.Resize(fyne.NewSize(chess.MainWindowSize, chess.MainWindowSize))
	chess.Dots = []Dot{}
	for i := 0; i < chess.BoardSize; i++ {
		for j := 0; j < chess.BoardSize; j++ {
			chess.Dots = append(chess.Dots, NewDot(i, j))
		}
	}
	chess.FullBoard = make(Board)
	for i := 0; i < chess.BoardSize; i++ {
		for j := 0; j < chess.BoardSize; j++ {
			d := NewDot(i, j)
			if i+1 < chess.BoardSize {
				chess.FullBoard[NewEdge(d, NewDot(i+1, j))] = struct{}{}
			}
			if j+1 < chess.BoardSize {
				chess.FullBoard[NewEdge(d, NewDot(i, j+1))] = struct{}{}
			}
		}
	}
	chess.EdgesCount = len(chess.FullBoard)
	chess.Boxes = []Box{}
	for _, d := range chess.Dots {
		if d.X() < chess.BoardSize-1 && d.Y() < chess.BoardSize-1 {
			chess.Boxes = append(chess.Boxes, Box(d))
		}
	}
	chess.EdgeNearBoxes = make(map[Edge][]Box)
	for e := range chess.FullBoard {
		x := e.Dot2().X() - 1
		y := e.Dot2().Y() - 1
		if x >= 0 && y >= 0 {
			chess.EdgeNearBoxes[e] = []Box{Box(e.Dot1()), Box(NewDot(x, y))}
			continue
		}
		chess.EdgeNearBoxes[e] = []Box{Box(e.Dot1())}
	}
	chess.BoxEdges = make(map[Box][]Edge)
	for _, b := range chess.Boxes {
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
		chess.BoxEdges[b] = edges
	}
	chess.DotCanvases = make(map[Dot]*canvas.Circle)
	chess.EdgesCanvases = make(map[Edge]*canvas.Line)
	chess.BoxesCanvases = make(map[Box]*canvas.Rectangle)
	chess.EdgeButtons = make(map[Edge]*widget.Button)
	chess.BoxesFilledColor = make(map[Box]color.Color)
	chess.Container = container.NewWithoutLayout()
	chess.SignChan = make(chan struct{}, 1)
	chess.MoveRecords = []MoveRecord{}
	chess.NowTurn = Player1Turn
	chess.PlayerScore = map[Turn]int{Player1Turn: 0, Player2Turn: 0}
	chess.GlobalBoard = make(Board)
	for _, b := range chess.Boxes {
		chess.BoxesCanvases[b] = NewBoxCanvas(b)
		chess.Container.Add(chess.BoxesCanvases[b])
	}
	for e := range chess.FullBoard {
		chess.EdgesCanvases[e] = NewEdgeCanvas(e)
		chess.Container.Add(chess.EdgesCanvases[e])
		chess.EdgeButtons[e] = widget.NewButton("", func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if chess.AIPlayer1 && chess.NowTurn == Player1Turn {
				return
			} else if chess.AIPlayer2 && chess.NowTurn == Player2Turn {
				return
			}
			AddEdge(e)
		})
		size, pos := getEdgeButtonSizeAndPosition(e)
		chess.EdgeButtons[e].Resize(size)
		chess.EdgeButtons[e].Move(pos)
		chess.Container.Add(chess.EdgeButtons[e])
	}
	for _, d := range chess.Dots {
		chess.DotCanvases[d] = NewDotCanvas(d)
		chess.Container.Add(chess.DotCanvases[d])
	}
	go func() {
		notifySignChan()
		for range chess.SignChan {
			mu.Lock()
			AddEdge(GetBestEdge())
			Refresh()
			mu.Unlock()
		}
	}()
	MainWindow.SetContent(chess.Container)
}

func RestartWithCall(size int) {
	NewGame(size)
	SendMessage("Game Start! BoardSize: %d", chess.BoardSize)
}

func Tip(nowStep int, box Box) {
	boxesCanvas := chess.BoxesCanvases[box]
	defer func() {
		boxesCanvasLock.Lock()
		defer boxesCanvasLock.Unlock()
		boxesCanvas.FillColor = chess.BoxesFilledColor[box]
	}()
	ticker := time.NewTicker(AnimationStepTime)
	defer ticker.Stop()
	for {
		for i := 0; i <= AnimationSteps; i++ {
			if nowStep != len(chess.GlobalBoard) {
				return
			}
			t := float64(i) / float64(AnimationSteps)
			boxesCanvas.FillColor = interpolateColor(TipColor, GetThemeColor(), t)
			boxesCanvas.Refresh()
			<-ticker.C
		}
		for i := 0; i <= AnimationSteps; i++ {
			if nowStep != len(chess.GlobalBoard) {
				return
			}
			t := float64(i) / float64(AnimationSteps)
			boxesCanvas.FillColor = interpolateColor(GetThemeColor(), TipColor, t)
			boxesCanvas.Refresh()
			<-ticker.C
		}
	}
}

func StoreMoveRecord() error {
	RecordFileName := fmt.Sprintf("Dox-and-Boxes %s.json", time.Now().Format(time.DateTime))
	f, err := os.Create(filepath.Join(LogRecordFileDir, RecordFileName))
	if err != nil {
		return err
	}
	j, err := sonic.Marshal(chess.MoveRecords)
	if err != nil {
		return err
	}
	if _, err := f.Write(j); err != nil {
		return err
	}
	return nil
}

func AddEdge(e Edge) {
	if chess.BoardSize <= 1 {
		return
	}
	if _, c := chess.GlobalBoard[e]; c {
		return
	}
	if e == 0 {
		return
	}
	chess.MoveRecords = append(chess.MoveRecords, MoveRecord{
		TimeStamp:    time.Now().Format("2006-01-02 15:04:05"),
		Step:         len(chess.GlobalBoard),
		Player:       chess.NowTurn.ToString(),
		MoveEdge:     e,
		Player1Score: chess.PlayerScore[Player1Turn],
		Player2Score: chess.PlayerScore[Player2Turn],
	})
	nowStep := len(chess.GlobalBoard)
	obtainsBoxes := ObtainsBoxes(chess.GlobalBoard, e)
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
	defer chess.EdgeButtons[e].Hide()
	boxesCanvasLock.Lock()
	for _, box := range obtainsBoxes {
		playerFilledColor := GetPlayerFilledColor()
		chess.BoxesCanvases[box].FillColor = playerFilledColor
		chess.BoxesFilledColor[box] = playerFilledColor
	}
	boxesCanvasLock.Unlock()
	chess.EdgesCanvases[e].StrokeColor = GetPlayerHighlightColor()
	chess.PlayerScore[chess.NowTurn] += score
	if score == 0 {
		chess.NowTurn.Change()
	}
	chess.GlobalBoard[e] = struct{}{}
	nowStep++
	for _, box := range chess.Boxes {
		edgesCountInBox := EdgesCountInBox(chess.GlobalBoard, box)
		if edgesCountInBox == 3 {
			go Tip(nowStep, box)
		}
	}
	if nowStep == chess.EdgesCount {
		StoreMoveRecord()
		if chess.PlayerScore[Player1Turn] > chess.PlayerScore[Player2Turn] {
			SendMessage("Player1 Win!")
		} else if chess.PlayerScore[Player1Turn] < chess.PlayerScore[Player2Turn] {
			SendMessage("Player2 Win!")
		} else if chess.PlayerScore[Player1Turn] == chess.PlayerScore[Player2Turn] {
			SendMessage("Draw!")
		}
		if chess.AutoRestart {
			go func() {
				time.Sleep(2 * time.Second)
				RestartWithCall(chess.BoardSize)
			}()
		}
		return
	}
	notifySignChan()
	chess.UndoMenuItem.Disabled = false
}

func Recover(MoveRecord []MoveRecord) {
	if chess.MusicOn {
		chess.MusicOn = !chess.MusicOn
		defer func() { chess.MusicOn = !chess.MusicOn }()
	}
	NewGame(chess.BoardSize)
	for _, r := range MoveRecord {
		AddEdge(r.MoveEdge)
	}
}

func GetMessage(head string, value bool) string {
	if value {
		return head + " ON"
	} else {
		return head + " OFF"
	}
}

func main() {
	os.Mkdir(LogRecordFileDir, os.ModePerm)

	chess.RestartMenuItem = &fyne.MenuItem{
		Label: "Restart",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			RestartWithCall(chess.BoardSize)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	chess.PauseMenuItem = &fyne.MenuItem{
		Label:    "Pause",
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeySpace},
		Action: func() {
			defer Refresh()
			if !chess.PauseState {
				mu.Lock()
				chess.PauseMenuItem.Label = "Continue"
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
				chess.PauseMenuItem.Label = "Pause"
				SendMessage("Game Continue")
			}
			chess.PauseState = !chess.PauseState
		},
	}

	chess.ScoreMenuItem = &fyne.MenuItem{
		Label: "Score",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			SendMessage("Player1 Score: %d\nPlayer2 Score: %d\n", chess.PlayerScore[Player1Turn], chess.PlayerScore[Player2Turn])
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyS},
	}

	chess.AddBoardSizeMenuItem = &fyne.MenuItem{
		Label: "AddBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			RestartWithCall(chess.BoardSize + 1)
			if !chess.ReduceBoardSizeMenuItem.Disabled {
				chess.ReduceBoardSizeMenuItem.Disabled = false
			}
			if chess.BoardSize != DefaultBoardSize || chess.DotDistance != DefaultBoardSize {
				chess.ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyEqual},
	}

	chess.ReduceBoardSizeMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardSize",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if chess.BoardSize <= 1 {
				return
			}
			RestartWithCall(chess.BoardSize - 1)
			if chess.BoardSize <= 1 {
				chess.ResetBoardMenuItem.Disabled = true
			}
			if chess.BoardSize != DefaultBoardSize || chess.DotDistance != DefaultBoardSize {
				chess.ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyMinus},
	}

	chess.UndoMenuItem = &fyne.MenuItem{
		Label: "Undo",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			moveRecord := append([]MoveRecord{}, chess.MoveRecords...)
			if len(moveRecord) > 0 {
				r := moveRecord[len(moveRecord)-1]
				SendMessage("Undo Edge " + r.MoveEdge.ToString())
				moveRecord = moveRecord[:len(moveRecord)-1]
				Recover(moveRecord)
			}
			if len(moveRecord) == 0 {
				chess.UndoMenuItem.Disabled = true
			}
			if chess.BoardSize != DefaultBoardSize || chess.DotDistance != DefaultBoardSize {
				chess.ResetBoardMenuItem.Disabled = false
			}
		},
		Disabled: true,
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyZ},
	}

	chess.AddBoardWidthMenuItem = &fyne.MenuItem{
		Label: "AddBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			SetDotDistance(chess.DotDistance + 10)
			if chess.ReduceBoardWidthMenuItem.Disabled {
				chess.ReduceBoardWidthMenuItem.Disabled = true
			}
			if chess.BoardSize != DefaultBoardSize || chess.DotDistance != DefaultBoardSize {
				chess.ResetBoardMenuItem.Disabled = false
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyUp},
	}

	chess.ReduceBoardWidthMenuItem = &fyne.MenuItem{
		Label: "ReduceBoardWidth",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if chess.DotDistance-10 < MinDotSize {
				return
			}
			SetDotDistance(chess.DotDistance - 10)
			if chess.DotDistance-10 < MinDotSize {
				chess.ReduceBoardSizeMenuItem.Disabled = true
			}
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyDown},
	}

	chess.ResetBoardMenuItem = &fyne.MenuItem{
		Label: "ResetBoard",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if chess.DotDistance != DefaultDotDistance {
				SetDotDistance(DefaultDotDistance)
			}
			if chess.BoardSize != DefaultBoardSize {
				RestartWithCall(DefaultBoardSize)
			}
			chess.ResetBoardMenuItem.Disabled = true
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyR},
	}

	chess.AIPlayer1MenuItem = &fyne.MenuItem{
		Label: GetMessage("AIPlayer1", !chess.AIPlayer1),
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if !chess.AIPlayer1 {
				notifySignChan()
			}
			message := GetMessage("AIPlayer1", chess.AIPlayer1)
			chess.AIPlayer1MenuItem.Label = message
			SendMessage(message)
			chess.AIPlayer1 = !chess.AIPlayer1
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key1},
	}

	chess.AIPlayer2MenuItem = &fyne.MenuItem{
		Label: GetMessage("AIPlayer2", !chess.AIPlayer2),
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			if !chess.AIPlayer2 {
				notifySignChan()
			}
			message := GetMessage("AIPlayer2", chess.AIPlayer2)
			chess.AIPlayer2MenuItem.Label = message
			SendMessage(message)
			chess.AIPlayer2 = !chess.AIPlayer2
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.Key2},
	}

	chess.MusicMenuItem = &fyne.MenuItem{
		Label: GetMessage("Music", !chess.MusicOn),
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			message := GetMessage("Music", chess.MusicOn)
			chess.MusicMenuItem.Label = message
			SendMessage(message)
			chess.MusicOn = !chess.MusicOn
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyP},
	}

	chess.AutoRestartMenuItem = &fyne.MenuItem{
		Label: GetMessage("AutoRestart", !chess.AutoRestart),
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			defer Refresh()
			message := GetMessage("AutoRestart", chess.AutoRestart)
			if !chess.AutoRestart {
				if len(chess.GlobalBoard) == chess.EdgesCount {
					RestartWithCall(chess.BoardSize)
				}
			}
			chess.AutoRestartMenuItem.Label = message
			SendMessage(message)
			chess.AutoRestart = !chess.AutoRestart
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyA},
	}

	chess.QuitMenuItem = &fyne.MenuItem{
		Label: "Quit",
		Action: func() {
			mu.Lock()
			defer mu.Unlock()
			SendMessage("Game Closed")
			Refresh()
			os.Exit(0)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyQ},
		IsQuit:   true,
	}

	chess.HelpMenuItem = &fyne.MenuItem{
		Label: "Help",
		Action: func() {
			SendMessage(HelpDoc)
		},
		Shortcut: &desktop.CustomShortcut{KeyName: fyne.KeyH},
	}

	MainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Game",
			chess.RestartMenuItem,
			chess.UndoMenuItem,
			chess.PauseMenuItem,
			chess.ScoreMenuItem,
			fyne.NewMenuItemSeparator(),
			chess.QuitMenuItem,
		),
		fyne.NewMenu("Board",
			chess.AddBoardWidthMenuItem,
			chess.ReduceBoardWidthMenuItem,
			chess.AddBoardSizeMenuItem,
			chess.ReduceBoardSizeMenuItem,
			fyne.NewMenuItemSeparator(),
			chess.ResetBoardMenuItem,
		),
		fyne.NewMenu("Config",
			chess.AIPlayer1MenuItem,
			chess.AIPlayer2MenuItem,
			chess.AutoRestartMenuItem,
			chess.MusicMenuItem,
		),
		fyne.NewMenu("Help",
			chess.HelpMenuItem,
		),
	))
	mu.Lock()
	MoveRecords := append([]MoveRecord{}, chess.MoveRecords...)
	SetDotDistance(chess.DotDistance)
	MainWindow.SetFixedSize(true)
	App.Settings().SetTheme(GameTheme{})
	go func() {
		time.Sleep(300 * time.Millisecond)
		if len(MoveRecords) > 0 {
			Recover(MoveRecords)
		} else {
			RestartWithCall(chess.BoardSize)
		}
		chess.Container.Refresh()
		mu.Unlock()
	}()
	MainWindow.ShowAndRun()
}
