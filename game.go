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
	"fyne.io/fyne/v2/widget"
	"github.com/bytedance/sonic"
)

// GameManager interface defines the core functions needed to manage the game state.
type GameManager interface {
	Restart(int)          // Restart the game with a new board size
	Recover([]MoveRecord) // Recover the game state from a list of move records
	AddEdge(Edge)         // Add an edge to the board
	Undo()                // Undo the last move
	Refresh()             // Refresh the game state and UI
	StartAIPlayer1()      // Start AI player 1
	StartAIPlayer2()      // Start AI player 2
}

// Instantiate the game manager
var game GameManager = gameManager{}

type gameManager struct{}

// Refresh updates the UI and saves the game state to a file.
func (g gameManager) Refresh() {
	RefreshMenu()
	Container.Refresh()
	j, err := sonic.Marshal(chess)
	if err != nil {
		SendMessage(err.Error())
	}
	if err := os.WriteFile(ChessMetaFileName, j, os.ModePerm); err != nil {
		SendMessage(err.Error())
	}
}

// notifySignChan sends a signal to the AI player's channel if it's their turn.
func (g gameManager) notifySignChan() {
	if (chess.AIPlayer1 && CurrentTurn == Player1Turn) || (chess.AIPlayer2 && CurrentTurn == Player2Turn) {
		select {
		case SignChan <- struct{}{}:
		default:
		}
	}
}

// restart initializes a new game with the specified board size.
func (g gameManager) restart(NewBoardSize int) {
	chess.BoardSize = NewBoardSize
	chess.BoardSizePower = Dot(chess.BoardSize * chess.BoardSize)
	chess.MainWindowSize = chess.DotCanvasDistance*float32(chess.BoardSize) + chess.BoardMargin - 5
	MainWindow.Resize(fyne.NewSize(chess.MainWindowSize, chess.MainWindowSize))

	// Initialize dots
	AllDots = []Dot{}
	for i := 0; i < chess.BoardSize; i++ {
		for j := 0; j < chess.BoardSize; j++ {
			AllDots = append(AllDots, NewDot(i, j))
		}
	}

	// Initialize edges
	AllEdges = make(map[Edge]struct{})
	for i := 0; i < chess.BoardSize; i++ {
		for j := 0; j < chess.BoardSize; j++ {
			d := NewDot(i, j)
			if i+1 < chess.BoardSize {
				AllEdges[NewEdge(d, NewDot(i+1, j))] = struct{}{}
			}
			if j+1 < chess.BoardSize {
				AllEdges[NewEdge(d, NewDot(i, j+1))] = struct{}{}
			}
		}
	}
	AllEdgesCount = len(AllEdges)

	// Initialize boxes
	AllBoxes = []Box{}
	for _, d := range AllDots {
		if d.X() < chess.BoardSize-1 && d.Y() < chess.BoardSize-1 {
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
	chess.ChessMoveRecords = []MoveRecord{}
	CurrentTurn = Player1Turn
	Player1Score = 0
	Player2Score = 0
	CurrentBoard = NewBoard()

	// Add boxes to the container
	boxesCanvasLock.Lock()
	for _, b := range AllBoxes {
		BoxesCanvases[b] = NewBoxCanvas(b)
		Container.Add(BoxesCanvases[b])
	}
	boxesCanvasLock.Unlock()

	// Add edges to the container
	for e := range AllEdges {
		EdgesCanvases[e] = NewEdgeCanvas(e)
		Container.Add(EdgesCanvases[e])
		EdgeButtons[e] = widget.NewButton("", func() {
			globalLock.Lock()
			defer globalLock.Unlock()
			if (chess.AIPlayer1 && CurrentTurn == Player1Turn) || (chess.AIPlayer2 && CurrentTurn == Player2Turn) {
				return
			}
			g.AddEdge(e)
			g.Refresh()
		})
		size, pos := getEdgeButtonSizeAndPosition(e)
		EdgeButtons[e].Resize(size)
		EdgeButtons[e].Move(pos)
		Container.Add(EdgeButtons[e])
	}

	// Add dots to the container
	for _, d := range AllDots {
		DotCanvases[d] = NewDotCanvas(d)
		Container.Add(DotCanvases[d])
	}

	// Start a goroutine to handle AI moves
	go func() {
		g.notifySignChan()
		for range SignChan {
			globalLock.Lock()
			bestEdge := searchEngine.GetBestEdge()
			g.AddEdge(bestEdge)
			g.Refresh()
			globalLock.Unlock()
		}
	}()
	MainWindow.SetContent(Container)
}

// Restart restarts the game with the given board size and sends a message.
func (g gameManager) Restart(size int) {
	g.restart(size)
	SendMessage("Game Start! BoardSize: %v", chess.BoardSize)
}

// storeMoveRecord saves the current game state to a log file.
func (g gameManager) storeMoveRecord(WinMessage string) {
	startTimeStamp := chess.ChessMoveRecords[0].TimeStamp.Format(time.DateTime)
	endTimeStamp := chess.ChessMoveRecords[len(chess.ChessMoveRecords)-1].TimeStamp.Format(time.DateTime)
	gameName := fmt.Sprintf("Game %v", startTimeStamp)
	f, err := os.Create(gameName + ".log")
	if err != nil {
		SendMessage(err.Error())
		return
	}
	record := fmt.Sprintf("%v BoardSize: %v\n", startTimeStamp, chess.BoardSize)
	for _, r := range chess.ChessMoveRecords {
		record = record + r.String() + "\n"
	}
	record += endTimeStamp + " " + WinMessage
	if _, err := f.WriteString(record); err != nil {
		SendMessage(err.Error())
		return
	}
}

// startTipAnimation starts an animation to highlight a potential move.
func (g gameManager) startTipAnimation(nowStep int, boxesCanvas *canvas.Rectangle) {
	var animation *fyne.Animation
	currentThemeVariant := CurrentThemeVariant
	animation = canvas.NewColorRGBAAnimation(TipColor, gameTheme.GetThemeColor(), time.Second, func(c color.Color) {
		if nowStep != CurrentBoard.Size() {
			animation.Stop()
			return
		}
		if currentThemeVariant != CurrentThemeVariant {
			animation.Stop()
			g.startTipAnimation(nowStep, boxesCanvas)
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
func (g gameManager) AddEdge(e Edge) {
	if chess.BoardSize <= 1 {
		return
	}
	if CurrentBoard.Contains(e) {
		return
	}
	if e == InvalidEdge {
		return
	}
	chess.ChessMoveRecords = append(chess.ChessMoveRecords, MoveRecord{
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
	if chess.OpenMusic {
		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()
		go func() {
			defer wg.Done()
			if score > 0 {
				if err := musicPlayer.PlayScoreMusic(); err != nil {
					SendMessage(err.Error())
				}
			} else {
				if err := musicPlayer.PlayMoveMusic(); err != nil {
					SendMessage(err.Error())
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
				g.startTipAnimation(nowStep, boxesCanvas)
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
		SendMessage(WinMessage)
		g.storeMoveRecord(WinMessage)
		if chess.AutoRestartGame {
			go func() {
				time.Sleep(2 * time.Second)
				g.Restart(chess.BoardSize)
			}()
		}
	}
	g.notifySignChan()
}

// Undo reverts the last move.
func (g gameManager) Undo() {
	moveRecord := append([]MoveRecord{}, chess.ChessMoveRecords...)
	if len(moveRecord) > 0 {
		r := moveRecord[len(moveRecord)-1]
		SendMessage("Undo Edge %v", r.MoveEdge)
		moveRecord = moveRecord[:len(moveRecord)-1]
		g.Recover(moveRecord)
	}
}

// Recover replays the move records to restore the game state.
func (g gameManager) Recover(MoveRecord []MoveRecord) {
	if chess.OpenMusic {
		chess.OpenMusic = !chess.OpenMusic
		defer func() { chess.OpenMusic = !chess.OpenMusic }()
	}
	g.restart(chess.BoardSize)
	for _, r := range MoveRecord {
		g.AddEdge(r.MoveEdge)
	}
	chess.ChessMoveRecords = MoveRecord
}

// StartAIPlayer1 starts or stops AI player 1.
func (g gameManager) StartAIPlayer1() {
	if !chess.AIPlayer1 {
		g.notifySignChan()
	}
	message := GetMessage("AIPlayer1", !chess.AIPlayer1)
	SendMessage(message)
	chess.AIPlayer1 = !chess.AIPlayer1
}

// StartAIPlayer2 starts or stops AI player 2.
func (g gameManager) StartAIPlayer2() {
	if !chess.AIPlayer2 {
		g.notifySignChan()
	}
	message := GetMessage("AIPlayer2", !chess.AIPlayer2)
	SendMessage(message)
	chess.AIPlayer2 = !chess.AIPlayer2
}
