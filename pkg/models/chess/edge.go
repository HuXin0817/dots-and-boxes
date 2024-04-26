package chess

import "fmt"

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
