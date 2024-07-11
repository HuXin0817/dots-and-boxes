package main

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
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
	App                = app.New()
	MainWindow         = App.NewWindow("Dots and Boxes")
	BoardSize          int
	BoardSizePower     Dot
	DotDistance        float32
	DotWidth           float32
	DotMargin          float32
	BoxSize            float32
	MainWindowSize     float32
	GlobalThemeVariant fyne.ThemeVariant
	EdgesCount         int
	Dots               []Dot
	Boxes              []Box
	FullBoard          Board
	EdgeNearBoxes      map[Edge][]Box
	BoxEdges           map[Box][]Edge
	DotCanvases        map[Dot]*canvas.Circle
	EdgesCanvases      map[Edge]*canvas.Line
	BoxesCanvases      map[Box]*canvas.Rectangle
	EdgeButtons        map[Edge]*widget.Button
	BoxesFilledColor   map[Box]color.Color
	PlayerScore        map[Turn]int
	Container          *fyne.Container
	SignChan           chan struct{}
	MoveRecord         []Edge
	GlobalBoard        Board
	NowTurn            Turn
	mu                 sync.Mutex
	boxesCanvasMu      sync.Mutex
	AIPlayer1          = NewOption("AIPlayer1", false)
	AIPlayer2          = NewOption("AIPlayer2", false)
	AutoRestart        = NewOption("AutoRestart", false)
	MusicOn            = NewOption("Music", true)
)

func main() {
	ResetBoard()
	MainWindow.SetFixedSize(true)
	App.Settings().SetTheme(GameTheme{})
	go func() {
		time.Sleep(300 * time.Millisecond)
		Container.Refresh()
	}()
	MainWindow.ShowAndRun()
}
