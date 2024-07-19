package main

import (
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	// CurrentThemeVariant Global variables for current gameTheme variant and game state
	CurrentThemeVariant fyne.ThemeVariant

	LightThemeDotCanvasColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // #FFFFFF
	DarkThemeDotCanvasColor  = color.NRGBA{R: 0xCA, G: 0xCA, B: 0xCA, A: 0xFF} // #CACACA
	LightThemeColor          = color.NRGBA{R: 0xF2, G: 0xF2, B: 0xF2, A: 0xFF} // #F2F2F2
	DarkThemeColor           = color.NRGBA{R: 0x2B, G: 0x2B, B: 0x2B, A: 0xFF} // #2B2B2B
	LightThemeButtonColor    = color.NRGBA{R: 0xD9, G: 0xD9, B: 0xD9, A: 0xFF} // #D9D9D9
	DarkThemeButtonColor     = color.NRGBA{R: 0x41, G: 0x41, B: 0x41, A: 0xFF} // #414141
	Player1HighlightColor    = color.NRGBA{R: 0x40, G: 0x40, B: 0xFF, A: 0x80} // #4040FF80
	Player2HighlightColor    = color.NRGBA{R: 0xFF, G: 0x40, B: 0x40, A: 0x80} // #FF404080
	Player1FilledColor       = color.NRGBA{R: 0x40, G: 0x40, B: 0xFF, A: 0x40} // #4040FF40
	Player2FilledColor       = color.NRGBA{R: 0xFF, G: 0x40, B: 0x40, A: 0x40} // #FF404040
	TipColor                 = color.NRGBA{R: 0xFF, G: 0xFF, B: 0x40, A: 0x40} // #FFFF4040
)

// Theme implements the fyne.Theme interface
type Theme struct{}

var gameTheme = &Theme{}

// Color returns the color for a given gameTheme element and variant (light/dark)
func (g *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// If the gameTheme variant has changed, update the colors of the canvas elements
	if CurrentThemeVariant != variant {
		CurrentThemeVariant = variant
		// Update the color of dot canvases
		for _, circle := range DotCanvases {
			circle.FillColor = g.GetDotCanvasColor()
			circle.Refresh()
		}
		// Update the color of box canvases
		boxesCanvasLock.Lock()
		for box, rectangle := range BoxesCanvases {
			if _, c := BoxesFilledColor[box]; !c {
				rectangle.FillColor = g.GetThemeColor()
				rectangle.Refresh()
			}
		}
		boxesCanvasLock.Unlock()
	}

	// Return the appropriate color based on the element name
	switch name {
	case theme.ColorNameBackground:
		return g.GetThemeColor()
	case theme.ColorNameButton:
		return g.GetButtonColor()
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Icon returns the main window icon for the given gameTheme icon name
func (g *Theme) Icon(fyne.ThemeIconName) fyne.Resource { return MainWindow.Icon() }

// Font returns the font resource for the given text style
func (g *Theme) Font(style fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(style) }

// Size returns the size for the given gameTheme size name
func (g *Theme) Size(name fyne.ThemeSizeName) float32 { return theme.DefaultTheme().Size(name) }

// getColorByVariant returns the appropriate color based on the current gameTheme variant
func (g *Theme) getColorByVariant(lightColor, darkColor color.Color) color.Color {
	if CurrentThemeVariant == theme.VariantDark {
		return darkColor
	} else {
		return lightColor
	}
}

// GetDotCanvasColor returns the color for dot canvases based on the current gameTheme variant
func (g *Theme) GetDotCanvasColor() color.Color {
	return g.getColorByVariant(LightThemeDotCanvasColor, DarkThemeDotCanvasColor)
}

// GetThemeColor returns the general gameTheme color based on the current gameTheme variant
func (g *Theme) GetThemeColor() color.Color {
	return g.getColorByVariant(LightThemeColor, DarkThemeColor)
}

// GetButtonColor returns the button color based on the current gameTheme variant
func (g *Theme) GetButtonColor() color.Color {
	return g.getColorByVariant(LightThemeButtonColor, DarkThemeButtonColor)
}

// GetPlayerFilledColor returns the color used to fill boxes based on the current player's turn
func (g *Theme) GetPlayerFilledColor() color.Color {
	if CurrentTurn == Player1Turn {
		return Player1FilledColor
	} else {
		return Player2FilledColor
	}
}

// GetPlayerHighlightColor returns the color used to highlight moves based on the current player's turn
func (g *Theme) GetPlayerHighlightColor() color.Color {
	if CurrentTurn == Player1Turn {
		return Player1HighlightColor
	} else {
		return Player2HighlightColor
	}
}

// UI interface defines the core functions needed to manage the game state.
type UI interface {
	Restart(int)            // Restart the game with a new board size
	Recover([]MoveRecord)   // Recover the game state from a list of move records
	AddEdge(Edge)           // Add an edge to the board
	Undo()                  // Undo the last move
	Refresh()               // Refresh the game state and UI
	StartAIPlayer1()        // Start AI player 1
	StartAIPlayer2()        // Start AI player 2
	SetDotDistance(float32) // Set UI DotDistance
}

// Instantiate the game manager
var game UI = &ui{}

type ui struct{}

// transPosition translates a coordinate to its position on the canvas.
func (ui *ui) transPosition(x int) float32 {
	return Chess.BoardMargin + float32(x)*Chess.DotCanvasDistance
}

// GetDotPosition returns the position of a dot on the canvas.
func (ui *ui) GetDotPosition(d Dot) (float32, float32) {
	return ui.transPosition(d.X()), ui.transPosition(d.Y())
}

// getEdgeButtonSizeAndPosition calculates the size and position of the edge button.
func (ui *ui) getEdgeButtonSizeAndPosition(e Edge) (size fyne.Size, pos fyne.Position) {
	if e.Dot1().X() == e.Dot2().X() {
		size = fyne.NewSize(Chess.DotCanvasWidth, Chess.DotCanvasDistance)
		pos = fyne.NewPos(
			(ui.transPosition(e.Dot1().X())+ui.transPosition(e.Dot2().X()))/2-size.Width/2+Chess.DotCanvasWidth/2,
			(ui.transPosition(e.Dot1().Y())+ui.transPosition(e.Dot2().Y()))/2-size.Height/2+Chess.DotCanvasWidth/2,
		)
	} else {
		size = fyne.NewSize(Chess.DotCanvasDistance, Chess.DotCanvasWidth)
		pos = fyne.NewPos(
			(ui.transPosition(e.Dot1().X())+ui.transPosition(e.Dot2().X()))/2-size.Width/2+Chess.DotCanvasWidth/2,
			(ui.transPosition(e.Dot1().Y())+ui.transPosition(e.Dot2().Y()))/2-size.Height/2+Chess.DotCanvasWidth/2,
		)
	}
	return
}

// NewDotCanvas creates a new dot canvas for the specified dot.
func (ui *ui) NewDotCanvas(d Dot) *canvas.Circle {
	newDotCanvas := canvas.NewCircle(gameTheme.GetDotCanvasColor())
	newDotCanvas.Resize(fyne.NewSize(Chess.DotCanvasWidth, Chess.DotCanvasWidth))
	newDotCanvas.Move(fyne.NewPos(ui.GetDotPosition(d)))
	return newDotCanvas
}

// NewEdgeCanvas creates a new edge canvas for the specified edge.
func (ui *ui) NewEdgeCanvas(e Edge) *canvas.Line {
	x1 := ui.transPosition(e.Dot1().X()) + Chess.DotCanvasWidth/2
	y1 := ui.transPosition(e.Dot1().Y()) + Chess.DotCanvasWidth/2
	x2 := ui.transPosition(e.Dot2().X()) + Chess.DotCanvasWidth/2
	y2 := ui.transPosition(e.Dot2().Y()) + Chess.DotCanvasWidth/2
	newEdgeCanvas := canvas.NewLine(gameTheme.GetDotCanvasColor())
	newEdgeCanvas.Position1 = fyne.NewPos(x1, y1)
	newEdgeCanvas.Position2 = fyne.NewPos(x2, y2)
	newEdgeCanvas.StrokeWidth = Chess.DotCanvasWidth
	return newEdgeCanvas
}

// NewBoxCanvas creates a new box canvas for the specified box.
func (ui *ui) NewBoxCanvas(box Box) *canvas.Rectangle {
	d := Dot(box)
	x := ui.transPosition(d.X()) + Chess.DotCanvasWidth
	y := ui.transPosition(d.Y()) + Chess.DotCanvasWidth
	newBoxCanvas := canvas.NewRectangle(gameTheme.GetThemeColor())
	newBoxCanvas.Move(fyne.NewPos(x, y))
	newBoxCanvas.Resize(fyne.NewSize(Chess.BoxCanvasSize, Chess.BoxCanvasSize))
	return newBoxCanvas
}

// Refresh updates the UI and saves the game state to a file.
func (ui *ui) Refresh() {
	RefreshMenu()
	Container.Refresh()
	if err := Chess.Refresh(); err != nil {
		Message.Send(err.Error())
	}
}

// notifySignChan sends a signal to the AI player's channel if it's their turn.
func (ui *ui) notifySignChan() {
	if (Chess.AIPlayer1 && CurrentTurn == Player1Turn) || (Chess.AIPlayer2 && CurrentTurn == Player2Turn) {
		select {
		case SignChan <- struct{}{}:
		default:
		}
	}
}

// restart initializes a new game with the specified board size.
func (ui *ui) restart(NewBoardSize int) {
	Chess.BoardSize = NewBoardSize
	Chess.BoardSizePower = Dot(Chess.BoardSize * Chess.BoardSize)
	Chess.MainWindowSize = Chess.DotCanvasDistance*float32(Chess.BoardSize) + Chess.BoardMargin - 5
	MainWindow.Resize(fyne.NewSize(Chess.MainWindowSize, Chess.MainWindowSize))

	// Initialize dots
	AllDots = []Dot{}
	for i := 0; i < Chess.BoardSize; i++ {
		for j := 0; j < Chess.BoardSize; j++ {
			AllDots = append(AllDots, NewDot(i, j))
		}
	}

	// Initialize edges
	AllEdges = make(map[Edge]struct{})
	for i := 0; i < Chess.BoardSize; i++ {
		for j := 0; j < Chess.BoardSize; j++ {
			d := NewDot(i, j)
			if i+1 < Chess.BoardSize {
				AllEdges[NewEdge(d, NewDot(i+1, j))] = struct{}{}
			}
			if j+1 < Chess.BoardSize {
				AllEdges[NewEdge(d, NewDot(i, j+1))] = struct{}{}
			}
		}
	}
	AllEdgesCount = len(AllEdges)

	// Initialize boxes
	AllBoxes = []Box{}
	for _, d := range AllDots {
		if d.X() < Chess.BoardSize-1 && d.Y() < Chess.BoardSize-1 {
			AllBoxes = append(AllBoxes, Box(d))
		}
	}

	// Initialize edge-adjacent boxes
	EdgeAdjacentBoxes = make(map[Edge][]Box)
	for e := range AllEdges {
		x := e.Dot2().X() - 1
		y := e.Dot2().Y() - 1
		if x >= 0 && y >= 0 {
			EdgeAdjacentBoxes[e] = []Box{Box(e.Dot1()), Box(NewDot(x, y))}
			continue
		}
		EdgeAdjacentBoxes[e] = []Box{Box(e.Dot1())}
	}

	// Initialize all edges in each box
	AllEdgesInBox = make(map[Box][]Edge)
	for _, b := range AllBoxes {
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
		AllEdgesInBox[b] = edges
	}

	// Initialize canvases
	DotCanvases = make(map[Dot]*canvas.Circle)
	EdgesCanvases = make(map[Edge]*canvas.Line)
	boxesCanvasLock.Lock()
	BoxesCanvases = make(map[Box]*canvas.Rectangle)
	boxesCanvasLock.Unlock()
	EdgeButtons = make(map[Edge]*widget.Button)
	BoxesFilledColor = make(map[Box]color.Color)
	Container = container.NewWithoutLayout()
	Chess.ChessMoveRecords = []MoveRecord{}
	CurrentTurn = Player1Turn
	Player1Score = 0
	Player2Score = 0
	CurrentBoard = NewBoard()

	// Add boxes to the container
	boxesCanvasLock.Lock()
	for _, b := range AllBoxes {
		BoxesCanvases[b] = ui.NewBoxCanvas(b)
		Container.Add(BoxesCanvases[b])
	}
	boxesCanvasLock.Unlock()

	// Add edges to the container
	for e := range AllEdges {
		EdgesCanvases[e] = ui.NewEdgeCanvas(e)
		Container.Add(EdgesCanvases[e])
		EdgeButtons[e] = widget.NewButton("", func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			if (Chess.AIPlayer1 && CurrentTurn == Player1Turn) || (Chess.AIPlayer2 && CurrentTurn == Player2Turn) {
				return
			}
			ui.AddEdge(e)
			ui.Refresh()
		})
		size, pos := ui.getEdgeButtonSizeAndPosition(e)
		EdgeButtons[e].Resize(size)
		EdgeButtons[e].Move(pos)
		Container.Add(EdgeButtons[e])
	}

	// Add dots to the container
	for _, d := range AllDots {
		DotCanvases[d] = ui.NewDotCanvas(d)
		Container.Add(DotCanvases[d])
	}

	// Start a goroutine to handle AI moves
	go func() {
		ui.notifySignChan()
		for range SignChan {
			globalLock.Lock()
			ui.AddEdge(GetBestEdge())
			ui.Refresh()
			globalLock.Unlock()
		}
	}()
	MainWindow.SetContent(Container)
}

// Restart restarts the game with the given board size and sends a message.
func (ui *ui) Restart(size int) {
	ui.restart(size)
	Message.Send("Game Start! BoardSize: %v", Chess.BoardSize)
}

// storeMoveRecord saves the current game state to a log file.
func (ui *ui) storeMoveRecord(WinMessage string) {
	startTimeStamp := Chess.ChessMoveRecords[0].TimeStamp.Format(time.DateTime)
	endTimeStamp := Chess.ChessMoveRecords[len(Chess.ChessMoveRecords)-1].TimeStamp.Format(time.DateTime)
	gameName := fmt.Sprintf("Game %v", startTimeStamp)
	f, err := os.Create(gameName + ".log")
	if err != nil {
		Message.Send(err.Error())
		return
	}
	record := fmt.Sprintf("%v BoardSize: %v\n", startTimeStamp, Chess.BoardSize)
	for _, r := range Chess.ChessMoveRecords {
		record = record + r.String() + "\n"
	}
	record += endTimeStamp + " " + WinMessage
	if _, err := f.WriteString(record); err != nil {
		Message.Send(err.Error())
		return
	}
}

// startTipAnimation starts an animation to highlight a potential move.
func (ui *ui) startTipAnimation(nowStep int, boxesCanvas *canvas.Rectangle) {
	var animation *fyne.Animation
	currentThemeVariant := CurrentThemeVariant
	animation = canvas.NewColorRGBAAnimation(TipColor, gameTheme.GetThemeColor(), time.Second, func(c color.Color) {
		if nowStep != CurrentBoard.Size() {
			animation.Stop()
			return
		}
		if currentThemeVariant != CurrentThemeVariant {
			animation.Stop()
			ui.startTipAnimation(nowStep, boxesCanvas)
			return
		}
		boxesCanvas.FillColor = c
		boxesCanvas.Refresh()
	})
	animation.RepeatCount = fyne.AnimationRepeatForever
	animation.AutoReverse = true
	fyne.CurrentApp().Driver().StartAnimation(animation)
}

// AddEdge adds an edge to the board and updates the game state.
func (ui *ui) AddEdge(e Edge) {
	if Chess.BoardSize <= 1 {
		return
	}
	if CurrentBoard.Contains(e) {
		return
	}
	if e == InvalidEdge {
		return
	}
	Chess.ChessMoveRecords = append(Chess.ChessMoveRecords, MoveRecord{
		TimeStamp:    time.Now(),
		Step:         CurrentBoard.Size(),
		Player:       CurrentTurn,
		MoveEdge:     e,
		Player1Score: Player1Score,
		Player2Score: Player2Score,
	})
	nowStep := CurrentBoard.Size()
	obtainsBoxes := ObtainsBoxes(CurrentBoard, e)
	score := len(obtainsBoxes)
	if Chess.OpenMusic {
		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()
		go func() {
			defer wg.Done()
			if score > 0 {
				if err := musicPlayer.PlayScoreMusic(); err != nil {
					Message.Send(err.Error())
				}
			} else {
				if err := musicPlayer.PlayMoveMusic(); err != nil {
					Message.Send(err.Error())
				}
			}
		}()
	}
	EdgeButtons[e].Hide()
	boxesCanvasLock.Lock()
	for _, box := range obtainsBoxes {
		playerFilledColor := gameTheme.GetPlayerFilledColor()
		BoxesCanvases[box].FillColor = playerFilledColor
		BoxesFilledColor[box] = playerFilledColor
	}
	boxesCanvasLock.Unlock()
	EdgesCanvases[e].StrokeColor = gameTheme.GetPlayerHighlightColor()
	if CurrentTurn == Player1Turn {
		Player1Score += score
	} else {
		Player2Score += score
	}
	if score == 0 {
		ChangeTurn(&CurrentTurn)
	}
	CurrentBoard.Add(e)
	nowStep++
	for _, box := range AllBoxes {
		if EdgesCountInBox(CurrentBoard, box) == 3 {
			boxesCanvasLock.Lock()
			boxesCanvas := BoxesCanvases[box]
			originalColor := BoxesFilledColor[box]
			boxesCanvasLock.Unlock()
			go func() {
				ui.startTipAnimation(nowStep, boxesCanvas)
				boxesCanvasLock.Lock()
				boxesCanvas.FillColor = originalColor
				boxesCanvasLock.Unlock()
			}()
		}
	}
	if nowStep == AllEdgesCount {
		var WinMessage string
		if Player1Score > Player2Score {
			WinMessage = "Player1 Win!"
		} else if Player1Score < Player2Score {
			WinMessage = "Player2 Win!"
		} else if Player1Score == Player2Score {
			WinMessage = "Draw!"
		}
		Message.Send(WinMessage)
		ui.storeMoveRecord(WinMessage)
		if Chess.AutoRestartGame {
			go func() {
				time.Sleep(2 * time.Second)
				ui.Restart(Chess.BoardSize)
			}()
		}
	}
	ui.notifySignChan()
}

// Undo reverts the last move.
func (ui *ui) Undo() {
	moveRecord := append([]MoveRecord{}, Chess.ChessMoveRecords...)
	if len(moveRecord) > 0 {
		r := moveRecord[len(moveRecord)-1]
		Message.Send("Undo Edge %v", r.MoveEdge)
		moveRecord = moveRecord[:len(moveRecord)-1]
		ui.Recover(moveRecord)
	}
}

// Recover replays the move records to restore the game state.
func (ui *ui) Recover(MoveRecord []MoveRecord) {
	if Chess.OpenMusic {
		Chess.OpenMusic = !Chess.OpenMusic
		defer func() { Chess.OpenMusic = !Chess.OpenMusic }()
	}
	ui.restart(Chess.BoardSize)
	for _, r := range MoveRecord {
		ui.AddEdge(r.MoveEdge)
	}
	Chess.ChessMoveRecords = MoveRecord
}

// StartAIPlayer1 starts or stops AI player 1.
func (ui *ui) StartAIPlayer1() {
	if !Chess.AIPlayer1 {
		ui.notifySignChan()
	}
	message := GetMessage("AIPlayer1", !Chess.AIPlayer1)
	Message.Send(message)
	Chess.AIPlayer1 = !Chess.AIPlayer1
}

// StartAIPlayer2 starts or stops AI player 2.
func (ui *ui) StartAIPlayer2() {
	if !Chess.AIPlayer2 {
		ui.notifySignChan()
	}
	message := GetMessage("AIPlayer2", !Chess.AIPlayer2)
	Message.Send(message)
	Chess.AIPlayer2 = !Chess.AIPlayer2
}

// SetDotDistance sets the distance between dots and updates the board layout.
func (ui *ui) SetDotDistance(d float32) {
	Chess.DotCanvasDistance = d
	Chess.DotCanvasWidth = Chess.DotCanvasDistance / 5
	Chess.BoardMargin = Chess.DotCanvasDistance / 3 * 2
	Chess.BoxCanvasSize = Chess.DotCanvasDistance - Chess.DotCanvasWidth
	Chess.MainWindowSize = Chess.DotCanvasDistance*float32(Chess.BoardSize) + Chess.BoardMargin - 5
	MainWindow.Resize(fyne.NewSize(Chess.MainWindowSize, Chess.MainWindowSize))
	moveRecord := append([]MoveRecord{}, Chess.ChessMoveRecords...)
	game.Recover(moveRecord)
}
