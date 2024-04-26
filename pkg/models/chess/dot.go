package chess

const (
	D       = 8
	dotMod  = 1 << D
	dotMask = dotMod - 1
)

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
