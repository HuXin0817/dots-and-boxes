package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SearchTime = int(2e6)
	Goroutines = 36
)

func GetNextEdges(s Board) []Edge {
	canGetOneScoreEdges := make(map[Edge]struct{})
	mayGiveEnemyScoreEdges := make(map[int][]Edge)
	for _, e := range EdgesMap {
		if _, c := s[e]; !c {
			score := s.ObtainsScore(e)
			if score == 2 {
				return []Edge{e}
			} else if score == 1 {
				canGetOneScoreEdges[e] = struct{}{}
			} else if score == 0 {
				boxes := e.NearBoxes()
				enemyMaxCanGetScore := 0
				s[e] = struct{}{}
				for _, box := range boxes {
					if s.EdgesCountInBox(box) == 3 {
						enemyMaxCanGetScore++
					}
				}
				delete(s, e)
				mayGiveEnemyScoreEdges[enemyMaxCanGetScore] = append(mayGiveEnemyScoreEdges[enemyMaxCanGetScore], e)
			}
		}
	}
	if len(canGetOneScoreEdges) == 0 {
		if len(mayGiveEnemyScoreEdges[0]) > 0 {
			return mayGiveEnemyScoreEdges[0]
		}
		if len(mayGiveEnemyScoreEdges[1]) > 0 {
			return mayGiveEnemyScoreEdges[1]
		}
		if len(mayGiveEnemyScoreEdges[2]) > 0 {
			return mayGiveEnemyScoreEdges[1]
		}
	} else {
		if len(mayGiveEnemyScoreEdges[0]) > 0 {
			for e := range canGetOneScoreEdges {
				return []Edge{e}
			}
		}
		if len(mayGiveEnemyScoreEdges[1]) > 0 {
			var BetterEdges []Edge
			for e := range canGetOneScoreEdges {
				BetterEdges = append(BetterEdges, e)
			}
			for _, e := range mayGiveEnemyScoreEdges[1] {
				BetterEdges = append(BetterEdges, e)
			}
			return BetterEdges
		}
		var BetterEdges []Edge
		for e := range canGetOneScoreEdges {
			BetterEdges = append(BetterEdges, e)
		}
		return BetterEdges
	}
	return nil
}

func GenerateBestEdge(board Board) (bestEdge Edge) {
	nextEdges := GetNextEdges(board)
	nextEdgesLen := len(nextEdges)
	if nextEdgesLen == 1 {
		return nextEdges[0]
	}

	var mu sync.Mutex
	sumScore := make(map[Edge]int, nextEdgesLen)
	searchTime := make(map[Edge]int, nextEdgesLen)
	for _, e := range nextEdges {
		searchTime[e] = 0
		sumScore[e] = 0
	}

	var t atomic.Int64
	var wg sync.WaitGroup
	wg.Add(Goroutines)
	for range Goroutines {
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for int(t.Load()) < SearchTime {
				b := NewBoard(board)
				turn := Player1Turn
				var firstEdge Edge
				var score int
				for {
					next := GetNextEdges(b)
					if len(next) == 0 {
						break
					}
					e := next[r.Intn(len(next))]
					if firstEdge == 0 {
						firstEdge = e
					}
					s := b.ObtainsScore(e)
					score += int(turn) * s
					if s == 0 {
						turn.Change()
					}
					b[e] = struct{}{}
					t.Add(1)
				}

				mu.Lock()
				sumScore[firstEdge] += score
				searchTime[firstEdge]++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	bestScore := -1e9
	for e := range searchTime {
		avgScore := float64(sumScore[e]) / float64(searchTime[e])
		if avgScore > bestScore {
			bestEdge = e
			bestScore = avgScore
		}
	}
	return
}
