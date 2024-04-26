package chess

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
