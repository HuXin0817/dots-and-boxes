package assess

import (
	"math/rand"
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
)

func RandEdgeInBetterEdges(b chess.Board) chess.Edge {
	Edges := BetterEdges(b)
	return Edges[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(Edges))]
}

func BetterEdges(b chess.Board) []chess.Edge {
	ScoreCount := make(map[int][]chess.Edge)
	Edges := b.FreeEdges()
	for _, e := range Edges {
		Score := len(b.ObtainsBoxes(e))
		ScoreCount[Score] = append(ScoreCount[Score], e)
	}

	if len(ScoreCount[2]) > 0 {
		return ScoreCount[2]
	}

	if len(ScoreCount[1]) > 0 {
		return ScoreCount[1]
	}

	edgeSet := make(map[chess.Edge]struct{})
	for _, e := range ScoreCount[0] {
		edgeSet[e] = struct{}{}
	}

	for _, e := range ScoreCount[0] {
		boxes := e.NearBoxes()
		for _, box := range boxes {
			if len(b.Append(e).EdgesInBox(box)) == 3 {
				delete(edgeSet, e)
				break
			}
		}
	}

	if len(edgeSet) > 0 {
		var betterEdges []chess.Edge
		for e := range edgeSet {
			betterEdges = append(betterEdges, e)
		}
		return betterEdges
	}

	return ScoreCount[0]
}

func NextMoves(m Move) (moves []Move) {
	appendBoard := m.Board.Append(m.Edge)
	edges := BetterEdges(m.Board)

	for _, e := range edges {
		moves = append(moves, Move{
			Board: appendBoard,
			Edge:  e,
		})
	}

	return moves
}
