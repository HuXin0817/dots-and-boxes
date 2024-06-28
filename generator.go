package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SearchTime = int(2e6)
	Goroutines = int(32)
)

func GetNextEdges(s Board) map[Edge]struct{} {
	canGetOneScoreEdges := make(map[Edge]struct{})
	mayGiveEnemyScoreEdges := make(map[int]map[Edge]struct{})
	for i := range 3 {
		mayGiveEnemyScoreEdges[i] = make(map[Edge]struct{})
	}

	for _, e := range EdgesMap {
		if _, c := s[e]; !c {
			score := s.ObtainsScore(e)
			if score == 2 {
				return map[Edge]struct{}{e: {}}
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
				mayGiveEnemyScoreEdges[enemyMaxCanGetScore][e] = struct{}{}
			}
		}
	}

	if len(canGetOneScoreEdges) == 0 {
		if len(mayGiveEnemyScoreEdges[0]) > 0 {
			for e := range mayGiveEnemyScoreEdges[0] {
				return map[Edge]struct{}{e: {}}
			}
		}
		if len(mayGiveEnemyScoreEdges[1]) > 0 {
			return mayGiveEnemyScoreEdges[1]
		}
		if len(mayGiveEnemyScoreEdges[2]) > 0 {
			return mayGiveEnemyScoreEdges[2]
		}
	} else {
		if len(mayGiveEnemyScoreEdges[0]) > 0 {
			for e := range canGetOneScoreEdges {
				return map[Edge]struct{}{e: {}}
			}
		}
		if len(mayGiveEnemyScoreEdges[1]) > 0 {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			canGetOneScoreEdgesLen := float64(len(canGetOneScoreEdges)) * 10.0
			mayGiveEnemyScoreEdgesLen := float64(len(mayGiveEnemyScoreEdges[1]))

			if r.Float64() < canGetOneScoreEdgesLen/(canGetOneScoreEdgesLen+mayGiveEnemyScoreEdgesLen) {
				return canGetOneScoreEdges
			} else {
				return mayGiveEnemyScoreEdges[1]
			}
		}
		return canGetOneScoreEdges
	}
	return nil
}

func GetRandNextEdge(s Board) Edge {
	edges := GetNextEdges(s)
	for e := range edges {
		return e
	}
	panic("func GetRandNextEdge(s Board) Edge Fail")
}

func GenerateBestEdge(board Board) (bestEdge Edge) {
	nextEdges := GetNextEdges(board)
	if len(nextEdges) == 1 {
		for e := range nextEdges {
			return e
		}
	}

	var mu sync.Mutex
	sumScore := make(map[Edge]int, len(nextEdges))
	searchTime := make(map[Edge]int, len(nextEdges))
	for e := range nextEdges {
		searchTime[e] = 0
		sumScore[e] = 0
	}

	var t atomic.Int64
	var wg sync.WaitGroup
	wg.Add(Goroutines)
	for range Goroutines {
		go func() {
			defer wg.Done()
			for int(t.Load()) < SearchTime {
				b := NewBoard(board)
				turn := Player1Turn
				var firstEdge Edge
				var score int

				for len(b) < len(EdgesMap) {
					t.Add(1)
					edge := GetRandNextEdge(b)
					if firstEdge == 0 {
						firstEdge = edge
					}
					s := b.ObtainsScore(edge)
					score += int(turn) * s
					if s == 0 {
						turn.Change()
					}
					b[edge] = struct{}{}
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
