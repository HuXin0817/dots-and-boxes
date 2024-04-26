package chess

type Board struct {
	BoardSize int
	Edges     map[Edge]struct{}
}

func NewBoard(BoardSize int, edges ...Edge) (newBoard Board) {
	newBoard = Board{
		BoardSize: BoardSize,
		Edges:     make(map[Edge]struct{}),
	}

	for _, e := range edges {
		newBoard.Edges[e] = struct{}{}
	}

	return
}

func (b Board) EdgesInBox(box Box) (edges []Edge) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b.Edges[e]; c {
			edges = append(edges, e)
		}
	}
	return
}

func (b Board) FreeEdgesCount() int {
	return len(Edges(b.BoardSize)) - len(b.Edges)
}

func (b Board) FreeEdges() (freeEdges []Edge) {
	allEdges := Edges(b.BoardSize)
	for _, e := range allEdges {
		if _, c := b.Edges[e]; !c {
			freeEdges = append(freeEdges, e)
		}
	}
	return
}

func (b Board) ScoreCount() (scoreCount int) {
	allBoxes := Boxes(b.BoardSize)
	for _, box := range allBoxes {
		if len(b.EdgesInBox(box)) == 4 {
			scoreCount++
		}
	}

	return
}

func (b Board) RemainingScore() (remainingScore int) {
	allBoxes := Boxes(b.BoardSize)
	for _, box := range allBoxes {
		if len(b.EdgesInBox(box)) != 4 {
			remainingScore++
		}
	}

	return
}

func (b Board) Boxes() (boxes []Box) {
	allBoxes := Boxes(b.BoardSize)
	for _, box := range allBoxes {
		if len(b.EdgesInBox(box)) == 4 {
			boxes = append(boxes, box)
		}
	}

	return
}

func (b Board) DeepCopy() (newBoard Board) {
	edges := []Edge{}
	for e := range b.Edges {
		edges = append(edges, e)
	}
	return NewBoard(b.BoardSize, edges...)
}

func (b Board) Append(e Edge) (newBoard Board) {
	edges := []Edge{e}
	for e := range b.Edges {
		edges = append(edges, e)
	}
	return NewBoard(b.BoardSize, edges...)
}

func (b Board) ObtainsBoxes(e Edge) (obtainsBoxes []Box) {
	if _, c := b.Edges[e]; c {
		return
	}

	boxes := e.NearBoxes()
	for _, box := range boxes {
		if len(b.EdgesInBox(box)) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}
	return
}
