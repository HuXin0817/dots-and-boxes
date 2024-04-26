package assess

import (
	"math"
	"sync"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
)

const (
	MaxSearchTimes = 1e6
	INF            = float64(math.MaxInt64)
)

func Assess(m Move) float64 {
	return DFS(1, MaxSearchTimes, m, chess.Player1, 0, -INF, INF)
}

func DFS(depth int, searchTime int, move Move, turn chess.Turn, nowScore, alpha, beta float64) (score float64) {
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

	if turn == chess.Player1 {
		score = -INF
	} else {
		score = INF
	}

	if searchTime < MaxSearchTimes/10 {
		for _, m := range nextMoves {
			eval := DFS(depth+1, nextSearchTime, m, nextTurn, nowScore, alpha, beta)
			if turn == chess.Player1 {
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
			nextMovesEval[i] = DFS(depth+1, nextSearchTime, m, nextTurn, nowScore, alpha, beta)
			wg.Done()
		}(i, m)
	}
	wg.Wait()

	for _, eval := range nextMovesEval {
		if turn == chess.Player1 {
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
