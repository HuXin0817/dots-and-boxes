package main

import (
	"math"
	"sync"
)

const (
	MaxSearchTimes = 5e6
	INF            = float64(math.MaxInt64)
)

type Move struct {
	Board
	Edge
}

func (m Move) Score() float64 {
	return float64(len(m.Board.ObtainsBoxes(m.Edge)))
}

func (m Move) WillChangeTurn() bool {
	return len(m.Board.ObtainsBoxes(m.Edge)) == 0
}

func BetterEdges(b Board) []Edge {
	ScoreCount := make(map[int][]Edge)
	FreeEdges := b.FreeEdges()
	for _, e := range FreeEdges {
		Score := len(b.ObtainsBoxes(e))
		ScoreCount[Score] = append(ScoreCount[Score], e)
	}

	if len(ScoreCount[2]) > 0 {
		return ScoreCount[2]
	}

	if len(ScoreCount[1]) > 0 {
		return ScoreCount[1]
	}

	edgeSet := make(map[Edge]struct{})
	for _, e := range ScoreCount[0] {
		edgeSet[e] = struct{}{}
	}

	for _, e := range ScoreCount[0] {
		boxes := e.NearBoxes()
		for _, box := range boxes {
			if b.Append(e).EdgesCountInBox(box) == 3 {
				delete(edgeSet, e)
				break
			}
		}
	}

	if len(edgeSet) > 0 {
		var betterEdges []Edge
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

func DFS(depth int, searchTime int, move Move, turn Turn, nowScore, alpha, beta float64) (score float64) {
	nowScore += move.Score() * float64(turn) / float64(depth)
	if searchTime <= 1 {
		return nowScore
	}

	nextTurn := turn
	if move.WillChangeTurn() {
		nextTurn = -nextTurn
	}

	nextMoves := NextMoves(move)
	l := len(nextMoves)
	if l == 0 {
		return nowScore
	}

	nextSearchTime := searchTime / l

	if turn == Player1 {
		score = -INF
	} else {
		score = INF
	}

	if searchTime < MaxSearchTimes/10 {
		for _, m := range nextMoves {
			eval := DFS(depth+1, nextSearchTime-1, m, nextTurn, nowScore, alpha, beta)
			if turn == Player1 {
				score = math.Max(score, eval)
				alpha = math.Max(score, alpha)
			} else {
				score = math.Min(score, eval)
				beta = math.Min(score, beta)
			}

			if beta-alpha < -1 {
				break
			}
		}

		return
	}

	nextMovesEval := make([]float64, l)

	var wg sync.WaitGroup
	wg.Add(l)
	for i, m := range nextMoves {
		go func(i int, m Move) {
			nextMovesEval[i] = DFS(depth+1, nextSearchTime-1, m, nextTurn, nowScore, alpha, beta)
			wg.Done()
		}(i, m)
	}
	wg.Wait()

	for _, eval := range nextMovesEval {
		if turn == Player1 {
			score = math.Max(score, eval)
			alpha = math.Max(score, alpha)
		} else {
			score = math.Min(score, eval)
			beta = math.Min(score, beta)
		}

		if beta-alpha < -1 {
			break
		}
	}

	return
}

func GetBestMove(b Board) Edge {
	bestScore := math.Inf(-1)
	var bestMove Move

	edges := BetterEdges(b)
	for _, e := range edges {
		m := Move{
			Board: b,
			Edge:  e,
		}

		if bestScore < DFS(1, MaxSearchTimes/len(edges), m, Player1, 0, -INF, INF) {
			bestScore = m.Score()
			bestMove = m
		}
	}

	return bestMove.Edge
}
