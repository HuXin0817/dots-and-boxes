package main

import (
	"sync"
	"time"
)

type SearchEngine struct{}

var searchEngine SearchEngine

// getNextEdges evaluates and selects the next best edge to draw on the board.
// It returns the edge that either immediately obtains a score or minimizes the opponent's potential score.
func (se SearchEngine) getNextEdges(b Board) (bestEdge Edge) {
	enemyMinScore := 3
	for e := range AllEdges {
		// Check if the edge is not a part of the board
		if !b.Contains(e) {
			// Check if drawing this edge obtains a score
			if score := ObtainsScore(b, e); score > 0 {
				return e // Immediately return if it obtains a score
			} else if score == 0 {
				// Evaluate the potential score the opponent might gain
				boxes := e.AdjacentBoxes()
				enemyScore := 0
				for _, box := range boxes {
					if EdgesCountInBox(b, box) == 2 {
						enemyScore++ // Increment if the opponent could score here
					}
				}
				// Select the edge that minimizes the appointment's Engine potential score
				if enemyMinScore > enemyScore {
					enemyMinScore = enemyScore
					bestEdge = e
				}
			}
		}
	}
	return
}

// GetBestEdge performs a multithreaded search to determine the best edge to draw.
// It uses multiple goroutines to simulate the game and gather statistics on edge performance.
func (se SearchEngine) GetBestEdge() (bestEdge Edge) {
	// Maps to store global search times and scores for each edge
	globalSearchTime := make(map[Edge]int)
	globalSumScore := make(map[Edge]int)
	// Slice of maps to store local search times and scores for each goroutine
	localSearchTimes := make([]map[Edge]int, chess.AISearchGoroutines)
	localSumScores := make([]map[Edge]int, chess.AISearchGoroutines)
	var wg sync.WaitGroup
	wg.Add(chess.AISearchGoroutines)
	// Launch multiple goroutines for parallel edge evaluation
	for i := 0; i < chess.AISearchGoroutines; i++ {
		localSearchTime := make(map[Edge]int)
		localSumScore := make(map[Edge]int)
		localSearchTimes[i] = localSearchTime
		localSumScores[i] = localSumScore
		go func() {
			defer wg.Done()
			// Set a timeout for each goroutine to limit search time
			timer := time.NewTimer(chess.AISearchTime)
			for {
				select {
				case <-timer.C:
					return // Exit when the context times out
				default:
					// Clone the current board state
					b := CurrentBoard.Clone()
					firstEdge := InvalidEdge
					score := 0
					turn := Player1Turn
					// Simulate the game until all edges are drawn
					for b.Size() < AllEdgesCount {
						edge := se.getNextEdges(b)
						if firstEdge == InvalidEdge {
							firstEdge = edge
						}
						s := ObtainsScore(b, edge)
						score += int(turn) * s
						if s == 0 {
							ChangeTurn(&turn)
						}
						b.Add(edge)
					}
					// Update local statistics for the first edge chosen
					localSearchTime[firstEdge]++
					localSumScore[firstEdge] += score
				}
			}
		}()
	}
	wg.Wait() // Wait for all goroutines to finish

	// Aggregate local statistics into global statistics
	for i := range chess.AISearchGoroutines {
		for e, s := range localSearchTimes[i] {
			globalSearchTime[e] += s
		}
		for e, s := range localSumScores[i] {
			globalSumScore[e] += s
		}
	}

	// Determine the best edge based on the highest average score
	bestScore := -1e9
	for e, score := range globalSumScore {
		averageScore := float64(score) / float64(globalSearchTime[e])
		if averageScore > bestScore {
			bestEdge = e
			bestScore = averageScore
		}
	}
	return
}
