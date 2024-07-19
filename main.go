package main

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

// HelpDocUrl Constants for various configurations
const HelpDocUrl = "https://github.com/HuXin0817/dots-and-boxes/blob/main/README.md"

const (
	ChessMetaFileName              = "meta.json"      // File name for storing Chess meta data
	DefaultDotDistance             = 80               // Default distance between dots
	DefaultBoardSize               = 6                // Default board size
	DefaultStepTime                = time.Second      // Default time for each AI step
	DefaultPerformanceAnalysisTime = 30 * time.Second // Default time for performance analysis
	MinDotSize                     = 60               // Minimum size for dots
	MinBoardSize                   = 1                // Minimum board size
)

// ChessMeta stores the configuration and state of the game
type ChessMeta struct {
	BoardSize               int           `json:"boardSize"`               // Size of the board
	BoardSizePower          Dot           `json:"boardSizePower"`          // Power of the board size (used for edge calculations)
	DotCanvasWidth          float32       `json:"dotCanvasWidth"`          // Width of the dot canvas
	BoardMargin             float32       `json:"boardMargin"`             // Margin of the board
	BoxCanvasSize           float32       `json:"boxCanvasSize"`           // Size of the box canvas
	MainWindowSize          float32       `json:"mainWindowSize"`          // Size of the main window
	DotCanvasDistance       float32       `json:"dotCanvasDistance"`       // Distance between dots
	AIPlayer1               bool          `json:"aiPlayer1"`               // Flag for AI Player 1
	AIPlayer2               bool          `json:"aiPlayer2"`               // Flag for AI Player 2
	AutoRestartGame         bool          `json:"autoRestartGame"`         // Flag for auto-restart game
	OpenMusic               bool          `json:"openMusic"`               // Flag for opening music
	AISearchGoroutines      int           `json:"aiSearchGoroutines"`      // Number of goroutines for AI search
	AISearchTime            time.Duration `json:"aiSearchTime"`            // Time duration for AI search
	PerformanceAnalysisTime time.Duration `json:"performanceAnalysisTime"` // Time duration for performance analysis
	ChessMoveRecords        []MoveRecord  `json:"chessMoveRecords"`        // Records of Chess moves
}

// NewChessMeta initializes ChessMeta by reading from a file or setting default values.
func NewChessMeta() *ChessMeta {
	// Read from the meta file if it exists
	if b, err := os.ReadFile(ChessMetaFileName); err == nil {
		c := new(ChessMeta)
		// Unmarshal the JSON data
		if err := sonic.Unmarshal(b, c); err == nil {
			return c
		}
	}
	// Return default values if file does not exist or unmarshal fails
	return &ChessMeta{
		BoardSize:               DefaultBoardSize,
		DotCanvasDistance:       DefaultDotDistance,
		OpenMusic:               true,
		AISearchTime:            DefaultStepTime,
		AISearchGoroutines:      runtime.NumCPU(),
		PerformanceAnalysisTime: DefaultPerformanceAnalysisTime,
	}
}

// Global variables and initialization
var (
	Chess             = NewChessMeta()                        // Initialize Chess meta data
	Message           = NewMessageManager()                   // Initialize MessageManager
	Player1Score      int                                     // Score of Player 1
	Player2Score      int                                     // Score of Player 2
	CurrentBoard      Board                                   // Current state of the board
	CurrentTurn       Turn                                    // Current turn of the game
	AllEdgesCount     int                                     // Total number of edges
	AllDots           []Dot                                   // All dots on the board
	AllBoxes          []Box                                   // All boxes on the board
	AllEdges          map[Edge]struct{}                       // All edges on the board
	EdgeAdjacentBoxes map[Edge][]Box                          // Adjacent boxes for each edge
	AllEdgesInBox     map[Box][]Edge                          // All edges in each box
	SignChan          = make(chan struct{}, 1)                // Channel for signaling AI moves
	MainWindow        = app.New().NewWindow("Dots and Boxes") // Main window of the application
	Container         *fyne.Container                         // Container for holding UI elements
	DotCanvases       map[Dot]*canvas.Circle                  // Canvases for dots
	EdgesCanvases     map[Edge]*canvas.Line                   // Canvases for edges
	BoxesCanvases     map[Box]*canvas.Rectangle               // Canvases for boxes
	EdgeButtons       map[Edge]*widget.Button                 // Buttons for edges
	BoxesFilledColor  map[Box]color.Color                     // Colors for filled boxes

	globalLock      sync.Mutex // Global mutex for synchronization
	boxesCanvasLock sync.Mutex // Mutex for box canvas synchronization

	OutputLogFile    *os.File          // File for output log
	RefreshMacOSIcon = func([]byte) {} // Function for refreshing macOS icon
)

// Turn represents the current player's turn
type Turn int

const (
	Player1Turn Turn = 1            // Constant for Player 1's turn
	Player2Turn      = -Player1Turn // Constant for Player 2's turn
)

// String returns the string representation of the current turn.
func (t Turn) String() string {
	if t == Player1Turn {
		return "Player1"
	} else {
		return "Player2"
	}
}

// ChangeTurn switches the current turn to the other player.
func ChangeTurn(t *Turn) { *t = -*t }

// Dot represents a dot on the board.
type Dot int

// NewDot creates a new dot based on x and y coordinates.
func NewDot(x, y int) Dot { return Dot(x*Chess.BoardSize + y) }

// X returns the x-coordinate of the dot.
func (d Dot) X() int { return int(d) / Chess.BoardSize }

// Y returns the y-coordinate of the dot.
func (d Dot) Y() int { return int(d) % Chess.BoardSize }

// String returns the string representation of the dot.
func (d Dot) String() string { return fmt.Sprintf("(%v, %v)", d.X(), d.Y()) }

// Edge represents an edge between two dots.
type Edge int

const InvalidEdge Edge = 0 // Constant for invalid edge

// NewEdge creates a new edge between two dots.
func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*Chess.BoardSizePower + Dot2) }

// Dot1 returns the first dot of the edge.
func (e Edge) Dot1() Dot { return Dot(e) / Chess.BoardSizePower }

// Dot2 returns the second dot of the edge.
func (e Edge) Dot2() Dot { return Dot(e) % Chess.BoardSizePower }

// String returns the string representation of the edge.
func (e Edge) String() string { return fmt.Sprintf("%v => %v", e.Dot1(), e.Dot2()) }

// AdjacentBoxes returns the boxes adjacent to the edge.
func (e Edge) AdjacentBoxes() []Box { return EdgeAdjacentBoxes[e] }

// Box represents a box on the board.
type Box int

// Edges returns the edges that form the box.
func (b Box) Edges() []Edge { return AllEdgesInBox[b] }

// Board interface defines the methods for managing the board state.
type Board interface {
	Add(e Edge)           // Adds an edge to the board
	Contains(e Edge) bool // Checks if an edge is on the board
	Clone() Board         // Clones the board
	Size() int            // Returns the size of the board
}

// board is a concrete implementation of the Board interface.
type board map[Edge]struct{}

// NewBoard creates a new, empty board.
func NewBoard() Board { return make(board) }

// Add adds an edge to the board.
func (b board) Add(e Edge) { b[e] = struct{}{} }

// Contains checks if an edge is on the board.
func (b board) Contains(e Edge) bool {
	_, ok := b[e]
	return ok
}

// Size returns the number of edges on the board.
func (b board) Size() int { return len(b) }

// Clone creates a copy of the board.
func (b board) Clone() Board {
	cb := make(board, b.Size())
	for e := range b {
		cb.Add(e)
	}
	return &cb
}

// MoveRecord records a move in the game.
type MoveRecord struct {
	TimeStamp    time.Time `json:"timeStamp"`    // The timestamp of the move
	Step         int       `json:"step"`         // The step number of the move
	Player       Turn      `json:"player"`       // The player who made the move
	MoveEdge     Edge      `json:"moveEdge"`     // The edge that was moved
	Player1Score int       `json:"player1Score"` // The score of Player 1 after the move
	Player2Score int       `json:"player2Score"` // The score of Player 2 after the move
}

// String returns the string representation of the move record.
func (m MoveRecord) String() string {
	return fmt.Sprintf("%v Step: %v, Turn: %v, Edge: %v, Player1Score: %v, Player2Score: %v", m.TimeStamp.Format(time.DateTime), m.Step, m.Player, m.MoveEdge, m.Player1Score, m.Player2Score)
}

// EdgesCountInBox counts how many edges in the specified box are already on the board.
func EdgesCountInBox(b Board, box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if b.Contains(e) {
			count++
		}
	}
	return
}

// ObtainsScore checks how many boxes would be completed by adding an edge.
func ObtainsScore(b Board, e Edge) (count int) {
	if b.Contains(e) {
		return
	}
	boxes := e.AdjacentBoxes()
	for _, box := range boxes {
		if EdgesCountInBox(b, box) == 3 {
			count++
		}
	}
	return
}

// ObtainsBoxes returns the boxes that would be completed by adding an edge.
func ObtainsBoxes(b Board, e Edge) (obtainsBoxes []Box) {
	if b.Contains(e) {
		return
	}
	boxes := e.AdjacentBoxes()
	for _, box := range boxes {
		if EdgesCountInBox(b, box) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}
	return
}

// SetDotDistance sets the distance between dots and updates the board layout.
func SetDotDistance(d float32) {
	Chess.DotCanvasDistance = d
	Chess.DotCanvasWidth = Chess.DotCanvasDistance / 5
	Chess.BoardMargin = Chess.DotCanvasDistance / 3 * 2
	Chess.BoxCanvasSize = Chess.DotCanvasDistance - Chess.DotCanvasWidth
	Chess.MainWindowSize = Chess.DotCanvasDistance*float32(Chess.BoardSize) + Chess.BoardMargin - 5
	MainWindow.Resize(fyne.NewSize(Chess.MainWindowSize, Chess.MainWindowSize))
	moveRecord := append([]MoveRecord{}, Chess.ChessMoveRecords...)
	game.Recover(moveRecord)
}

// transPosition translates a coordinate to its position on the canvas.
func transPosition(x int) float32 { return Chess.BoardMargin + float32(x)*Chess.DotCanvasDistance }

// GetDotPosition returns the position of a dot on the canvas.
func GetDotPosition(d Dot) (float32, float32) { return transPosition(d.X()), transPosition(d.Y()) }

// getEdgeButtonSizeAndPosition calculates the size and position of the edge button.
func getEdgeButtonSizeAndPosition(e Edge) (size fyne.Size, pos fyne.Position) {
	if e.Dot1().X() == e.Dot2().X() {
		size = fyne.NewSize(Chess.DotCanvasWidth, Chess.DotCanvasDistance)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+Chess.DotCanvasWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+Chess.DotCanvasWidth/2,
		)
	} else {
		size = fyne.NewSize(Chess.DotCanvasDistance, Chess.DotCanvasWidth)
		pos = fyne.NewPos(
			(transPosition(e.Dot1().X())+transPosition(e.Dot2().X()))/2-size.Width/2+Chess.DotCanvasWidth/2,
			(transPosition(e.Dot1().Y())+transPosition(e.Dot2().Y()))/2-size.Height/2+Chess.DotCanvasWidth/2,
		)
	}
	return
}

// NewDotCanvas creates a new dot canvas for the specified dot.
func NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(gameTheme.GetDotCanvasColor())
	newDotCanvas.Resize(fyne.NewSize(Chess.DotCanvasWidth, Chess.DotCanvasWidth))
	newDotCanvas.Move(fyne.NewPos(GetDotPosition(d)))
	return newDotCanvas
}

// NewEdgeCanvas creates a new edge canvas for the specified edge.
func NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := transPosition(e.Dot1().X()) + Chess.DotCanvasWidth/2
	y1 := transPosition(e.Dot1().Y()) + Chess.DotCanvasWidth/2
	x2 := transPosition(e.Dot2().X()) + Chess.DotCanvasWidth/2
	y2 := transPosition(e.Dot2().Y()) + Chess.DotCanvasWidth/2
	newEdgeCanvas := canvas.NewLine(gameTheme.GetDotCanvasColor())
	newEdgeCanvas.Position1 = fyne.NewPos(x1, y1)
	newEdgeCanvas.Position2 = fyne.NewPos(x2, y2)
	newEdgeCanvas.StrokeWidth = Chess.DotCanvasWidth
	return newEdgeCanvas
}

// NewBoxCanvas creates a new box canvas for the specified box.
func NewBoxCanvas(box Box) *canvas.Rectangle {
	d := Dot(box)
	x := transPosition(d.X()) + Chess.DotCanvasWidth
	y := transPosition(d.Y()) + Chess.DotCanvasWidth
	newBoxCanvas := canvas.NewRectangle(gameTheme.GetThemeColor())
	newBoxCanvas.Move(fyne.NewPos(x, y))
	newBoxCanvas.Resize(fyne.NewSize(Chess.BoxCanvasSize, Chess.BoxCanvasSize))
	return newBoxCanvas
}

// the main is the entry point of the application.
func main() {
	// Set the Gin framework to release mode.
	gin.SetMode(gin.ReleaseMode)

	// Handle system signals for graceful shutdown.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGPIPE)
		sig := <-sigChan
		Message.Send("Received signal: %v\n", sig)
		game.Refresh()
		os.Exit(0)
	}()

	globalLock.Lock()
	MoveRecords := append([]MoveRecord{}, Chess.ChessMoveRecords...)
	MainWindow.SetFixedSize(true)
	fyne.CurrentApp().Settings().SetTheme(gameTheme)
	fyne.CurrentApp().Lifecycle().SetOnStopped(game.Refresh)

	// Initialize the game board and UI elements.
	go func() {
		time.Sleep(300 * time.Millisecond)
		if Chess.DotCanvasDistance == 0 {
			Chess.DotCanvasDistance = DefaultDotDistance
		}
		if Chess.BoardSize == 0 {
			Chess.BoardSize = DefaultBoardSize
		}
		SetDotDistance(Chess.DotCanvasDistance)
		if len(MoveRecords) > 0 {
			game.Recover(MoveRecords)
		} else {
			game.Restart(Chess.BoardSize)
		}
		game.Refresh()
		globalLock.Unlock()

		// Update the window icon continuously.
		go func() {
			for {
				img := MainWindow.Canvas().Capture()
				buf := new(bytes.Buffer)
				if err := png.Encode(buf, img); err != nil {
					Message.Send(err.Error())
					continue
				}
				icon := fyne.NewStaticResource("Dots-and-Boxes", buf.Bytes())
				MainWindow.SetIcon(icon)
				RefreshMacOSIcon(buf.Bytes())
				runtime.Gosched()
			}
		}()

		// Handle AI move signals.
		go func() {
			for range SignChan {
				globalLock.Lock()
				game.AddEdge(searchEngine.GetBestEdge())
				game.Refresh()
				globalLock.Unlock()
			}
		}()
	}()

	MainWindow.ShowAndRun()
}
