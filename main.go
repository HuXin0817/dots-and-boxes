package main

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"log"
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
	OutputLogFileName              = "output.log"     // File name for storing output logs
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

func (chess *ChessMeta) Refresh() error {
	j, err := sonic.Marshal(chess)
	if err != nil {
		return err
	}
	if err := os.WriteFile(ChessMetaFileName, j, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Global variables and initialization
var (
	Message           = NewMessageManager()                   // Initialize MessageManager
	Chess             = NewChessMeta()                        // Initialize Chess meta data
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

func GetMessage(head string, value bool) string {
	if value {
		return head + " ON"
	} else {
		return head + " OFF"
	}
}

type MessageManager struct {
	mu   sync.Mutex // Mutex for sent message synchronization
	file *os.File
}

func NewMessageManager() *MessageManager {
	// init initializes the output log file and handles potential errors.
	f, err := os.OpenFile(OutputLogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return &MessageManager{file: f}
}

// Send sends a notification and logs the message.
func (m *MessageManager) Send(format string, a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.Printf(format+"\n", a...)
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Dots-And-Boxes",
		Content: fmt.Sprintf(format, a...),
	})
	if _, err := m.file.WriteString(time.Now().Format(time.DateTime) + " " + fmt.Sprintf(format, a...) + "\n"); err != nil {
		log.Println(err)
		return
	}
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

// getNextEdges evaluates and selects the next best edge to draw on the board.
// It returns the edge that either immediately obtains a score or minimizes the opponent's potential score.
func getNextEdges(b Board) (bestEdge Edge) {
	enemyMinScore := 3
	for e := range AllEdges {
		// Check if the edge is not a part of the board
		if !b.Contains(e) {
			// Check if drawing this edge obtains a score
			if score := ObtainsScore(b, e); score > 0 {
				return e // Immediately return if it obtains a score
			} else if score == 0 {
				// Evaluate the potential score the opponent might gain
				boxes := e.AdjacentBoxes()
				enemyScore := 0
				for _, box := range boxes {
					if EdgesCountInBox(b, box) == 2 {
						enemyScore++ // Increment if the opponent could score here
					}
				}
				// Select the edge that minimizes the appointment's Engine potential score
				if enemyMinScore > enemyScore {
					enemyMinScore = enemyScore
					bestEdge = e
				}
			}
		}
	}
	return
}

// GetBestEdge performs a multithreaded search to determine the best edge to draw.
// It uses multiple goroutines to simulate the game and gather statistics on edge performance.
func GetBestEdge() (bestEdge Edge) {
	// Maps to store global search times and scores for each edge
	globalSearchTime := make(map[Edge]int)
	globalSumScore := make(map[Edge]int)
	// Slice of maps to store local search times and scores for each goroutine
	localSearchTimes := make([]map[Edge]int, Chess.AISearchGoroutines)
	localSumScores := make([]map[Edge]int, Chess.AISearchGoroutines)
	var wg sync.WaitGroup
	wg.Add(Chess.AISearchGoroutines)
	// Launch multiple goroutines for parallel edge evaluation
	for i := 0; i < Chess.AISearchGoroutines; i++ {
		localSearchTime := make(map[Edge]int)
		localSumScore := make(map[Edge]int)
		localSearchTimes[i] = localSearchTime
		localSumScores[i] = localSumScore
		go func() {
			defer wg.Done()
			// Set a timeout for each goroutine to limit search time
			timer := time.NewTimer(Chess.AISearchTime)
			for {
				select {
				case <-timer.C:
					return // Exit when the context times out
				default:
					// Clone the current board state
					b := CurrentBoard.Clone()
					firstEdge := InvalidEdge
					score := 0
					turn := Player1Turn
					// Simulate the game until all edges are drawn
					for b.Size() < AllEdgesCount {
						edge := getNextEdges(b)
						if firstEdge == InvalidEdge {
							firstEdge = edge
						}
						s := ObtainsScore(b, edge)
						score += int(turn) * s
						if s == 0 {
							ChangeTurn(&turn)
						}
						b.Add(edge)
					}
					// Update local statistics for the first edge chosen
					localSearchTime[firstEdge]++
					localSumScore[firstEdge] += score
				}
			}
		}()
	}
	wg.Wait() // Wait for all goroutines to finish

	// Aggregate local statistics into global statistics
	for i := range Chess.AISearchGoroutines {
		for e, s := range localSearchTimes[i] {
			globalSearchTime[e] += s
		}
		for e, s := range localSumScores[i] {
			globalSumScore[e] += s
		}
	}

	// Determine the best edge based on the highest average score
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
		game.SetDotDistance(Chess.DotCanvasDistance)
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
				game.AddEdge(GetBestEdge())
				game.Refresh()
				globalLock.Unlock()
			}
		}()
	}()

	MainWindow.ShowAndRun()
}
