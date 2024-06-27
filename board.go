package main

import "fmt"

const (
	D       = 8
	dotMod  = 1 << D
	dotMask = dotMod - 1
)

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

type Dot int

func NewDot(x, y int) Dot {
	return Dot((x << D) + y)
}

func (d Dot) X() int {
	return int(d) >> D
}

func (d Dot) Y() int {
	return int(d) & dotMask
}

var dotsMap = make(map[int][]Dot)

func Dots(BoardSize int) (dots []Dot) {
	if res, c := dotsMap[BoardSize]; c {
		return res
	}

	for i := range BoardSize {
		for j := range BoardSize {
			dots = append(dots, NewDot(i, j))
		}
	}

	dotsMap[BoardSize] = dots
	return
}

const (
	E        = D << 1
	edgeMod  = 1 << E
	edgeMask = edgeMod - 1
)

type Edge int

func NewEdge(Dot1, Dot2 Dot) Edge {
	if Dot1 > Dot2 {
		Dot1, Dot2 = Dot2, Dot1
	}
	return Edge((Dot1 << E) + Dot2)
}

func (e Edge) Dot1() Dot {
	return Dot(e) >> E
}

func (e Edge) Dot2() Dot {
	return Dot(e) & edgeMask
}

func (e Edge) String() string {
	return fmt.Sprintf("(%d, %d) -> (%d, %d)", e.Dot1().X(), e.Dot1().Y(), e.Dot2().X(), e.Dot2().Y())
}

func (e Edge) NearBoxes() (nearBoxes []Box) {
	nearBoxes = append(nearBoxes, Box(e.Dot1()))

	x := e.Dot2().X() - 1
	y := e.Dot2().Y() - 1
	if x >= 0 && y >= 0 {
		nearBoxes = append(nearBoxes, Box(NewDot(x, y)))
	}
	return
}

var edgesMap = make(map[int][]Edge)

func Edges(BoardSize int) (edges []Edge) {
	if res, c := edgesMap[BoardSize]; c {
		return res
	}

	for i := range BoardSize {
		for j := range BoardSize {
			d := NewDot(i, j)
			if i+1 < BoardSize {
				edges = append(edges, NewEdge(d, NewDot(i+1, j)))
			}

			if j+1 < BoardSize {
				edges = append(edges, NewEdge(d, NewDot(i, j+1)))
			}
		}
	}

	edgesMap[BoardSize] = edges
	return
}

type Box Dot

func (b Box) Dots() [4]Dot {
	x := Dot(b).X()
	y := Dot(b).Y()

	return [...]Dot{
		NewDot(x, y),
		NewDot(x+1, y),
		NewDot(x, y+1),
		NewDot(x+1, y+1),
	}
}

func (b Box) Edges() [4]Edge {
	x := Dot(b).X()
	y := Dot(b).Y()

	D00 := NewDot(x, y)
	D10 := NewDot(x+1, y)
	D01 := NewDot(x, y+1)
	D11 := NewDot(x+1, y+1)

	return [...]Edge{
		NewEdge(D00, D01),
		NewEdge(D00, D10),
		NewEdge(D10, D11),
		NewEdge(D01, D11),
	}
}

var BoxesMap = make(map[int][]Box)

func Boxes(BoardSize int) (boxes []Box) {
	if res, c := BoxesMap[BoardSize]; c {
		return res
	}

	dots := Dots(BoardSize - 1)
	for _, d := range dots {
		boxes = append(boxes, Box(d))
	}

	BoxesMap[BoardSize] = boxes
	return
}

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

func (b Board) EdgesCountInBox(box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b.Edges[e]; c {
			count++
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

func (b Board) Append(edge Edge) (newBoard Board) {
	edges := []Edge{edge}
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
		if b.EdgesCountInBox(box) == 3 {
			obtainsBoxes = append(obtainsBoxes, box)
		}
	}

	return
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
