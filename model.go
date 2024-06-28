package main

import "fmt"

type Turn int

const (
	Player1Turn Turn = 1
	Player2Turn Turn = -1
)

func (t *Turn) Change() { *t = -*t }

type (
	Dot  int
	Box  int
	Edge int

	Board map[Edge]struct{}
)

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

var Dots = func() (Dots []Dot) {
	for i := range BoardSize {
		for j := range BoardSize {
			Dots = append(Dots, NewDot(i, j))
		}
	}
	return
}()

const BoardSizePower = BoardSize * BoardSize

func NewEdge(Dot1, Dot2 Dot) Edge {
	if Dot1 > Dot2 {
		Dot1, Dot2 = Dot2, Dot1
	}
	return Edge(Dot1*BoardSizePower + Dot2)
}

func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

func (e Edge) ToString() string {
	return fmt.Sprintf("(%d, %d) => (%d, %d)", e.Dot1().X(), e.Dot1().Y(), e.Dot2().X(), e.Dot2().Y())
}

func (e Edge) NearBoxes() []Box {
	x := e.Dot2().X() - 1
	y := e.Dot2().Y() - 1
	if x >= 0 && y >= 0 {
		return []Box{Box(e.Dot1()), Box(NewDot(x, y))}
	}
	return []Box{Box(e.Dot1())}
}

var EdgesMap = func() (Edges []Edge) {
	for i := range BoardSize {
		for j := range BoardSize {
			d := NewDot(i, j)
			if i+1 < BoardSize {
				Edges = append(Edges, NewEdge(d, NewDot(i+1, j)))
			}
			if j+1 < BoardSize {
				Edges = append(Edges, NewEdge(d, NewDot(i, j+1)))
			}
		}
	}
	return
}()

func (b Box) Edges() []Edge {
	x := Dot(b).X()
	y := Dot(b).Y()

	D00 := NewDot(x, y)
	D10 := NewDot(x+1, y)
	D01 := NewDot(x, y+1)
	D11 := NewDot(x+1, y+1)

	return []Edge{
		NewEdge(D00, D01),
		NewEdge(D00, D10),
		NewEdge(D10, D11),
		NewEdge(D01, D11),
	}
}

var Boxes = func() (Boxes []Box) {
	for _, d := range Dots {
		if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
			Boxes = append(Boxes, Box(d))
		}
	}
	return Boxes
}()

func NewBoard(board Board) Board {
	newBoard := make(Board)
	for e := range board {
		newBoard[e] = struct{}{}
	}
	return newBoard
}

func (b Board) EdgesCountInBox(box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b[e]; c {
			count++
		}
	}
	return
}

func (b Board) ObtainsScore(e Edge) (count int) {
	if _, c := b[e]; c {
		return
	}

	boxes := e.NearBoxes()
	for _, box := range boxes {
		if b.EdgesCountInBox(box) == 3 {
			count++
		}
	}

	return
}

func (b Board) ObtainsBoxes(e Edge) (obtainsBoxes []Box) {
	if _, c := b[e]; c {
		return
	}

	boxes := e.NearBoxes()
	for _, box := range boxes {
		if b.EdgesCountInBox(box) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}

	return
}
