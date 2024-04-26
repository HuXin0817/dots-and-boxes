package chess

type Turn int8

const (
	Player1 Turn = 1
	Player2 Turn = -1
)

func (t Turn) String() string {
	switch t {
	case Player1:
		return "Player1"
	case Player2:
		return "Player2"
	}
	return ""
}

type Game struct {
	Board
	Player1Score int
	Player2Score int
	NowPlayer    Turn
}

func NewGame(BoardSize int) *Game {
	return &Game{
		Board:     NewBoard(BoardSize),
		NowPlayer: Player1,
	}
}

func (g *Game) Add(e Edge) {
	score := len(g.Board.ObtainsBoxes(e))
	switch g.NowPlayer {
	case Player1:
		g.Player1Score += score
	case Player2:
		g.Player2Score += score
	}

	if score == 0 {
		g.NowPlayer = -g.NowPlayer
	}

	g.Board = g.Board.Append(e)
}

func (g *Game) StepCount() int {
	return len(g.Board.Edges)
}
