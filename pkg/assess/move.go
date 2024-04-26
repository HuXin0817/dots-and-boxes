package assess

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
)

type Move struct {
	chess.Board
	chess.Edge
}

func (m Move) Score() float64 {
	return float64(len(m.Board.ObtainsBoxes(m.Edge)))
}

func (m Move) WillChangeTurn() bool {
	return len(m.Board.ObtainsBoxes(m.Edge)) == 0
}
