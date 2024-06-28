package main

import "fmt"

// Turn 表示轮次，玩家1或玩家2
type Turn int

const (
	Player1Turn Turn = 1  // 玩家1的轮次
	Player2Turn Turn = -1 // 玩家2的轮次
)

// Change 切换当前玩家轮次
func (t *Turn) Change() { *t = -*t }

// Dot 表示点，Box 表示方块，Edge 表示边
type (
	Dot  int
	Box  int
	Edge int

	// Board 表示棋盘，每条边由Edge表示
	Board map[Edge]struct{}
)

// NewDot 创建一个新的点
func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

// X 获取点的X坐标
func (d Dot) X() int { return int(d) / BoardSize }

// Y 获取点的Y坐标
func (d Dot) Y() int { return int(d) % BoardSize }

// Dots 生成所有点的列表
var Dots = func() (Dots []Dot) {
	for i := range BoardSize {
		for j := range BoardSize {
			Dots = append(Dots, NewDot(i, j))
		}
	}
	return
}()

// BoardSizePower 是BoardSize的平方，用于计算边的索引
const BoardSizePower = BoardSize * BoardSize

// NewEdge 创建一条新的边
func NewEdge(Dot1, Dot2 Dot) Edge {
	if Dot1 > Dot2 {
		Dot1, Dot2 = Dot2, Dot1
	}
	return Edge(Dot1*BoardSizePower + Dot2)
}

// Dot1 获取边的第一个点
func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

// Dot2 获取边的第二个点
func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

// ToString 将边转换为字符串表示
func (e Edge) ToString() string {
	return fmt.Sprintf("(%d, %d) => (%d, %d)", e.Dot1().X(), e.Dot1().Y(), e.Dot2().X(), e.Dot2().Y())
}

// EdgeNearBoxes 缓存边附近的方块
var EdgeNearBoxes = make(map[Edge][]Box)

// NearBoxes 获取边附近的方块
func (e Edge) NearBoxes() []Box {
	if boxes, c := EdgeNearBoxes[e]; c {
		return boxes
	}

	x := e.Dot2().X() - 1
	y := e.Dot2().Y() - 1
	if x >= 0 && y >= 0 {
		boxes := []Box{Box(e.Dot1()), Box(NewDot(x, y))}
		EdgeNearBoxes[e] = boxes
		return boxes
	}

	boxes := []Box{Box(e.Dot1())}
	EdgeNearBoxes[e] = boxes
	return boxes
}

// EdgesMap 生成所有边的列表
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

// BoxEdges 缓存方块的边
var BoxEdges = make(map[Box][]Edge)

// Edges 获取方块的边
func (b Box) Edges() []Edge {
	if edges, c := BoxEdges[b]; c {
		return edges
	}

	x := Dot(b).X()
	y := Dot(b).Y()

	D00 := NewDot(x, y)
	D10 := NewDot(x+1, y)
	D01 := NewDot(x, y+1)
	D11 := NewDot(x+1, y+1)

	edges := []Edge{
		NewEdge(D00, D01),
		NewEdge(D00, D10),
		NewEdge(D10, D11),
		NewEdge(D01, D11),
	}

	BoxEdges[b] = edges
	return edges
}

// Boxes 生成所有方块的列表
var Boxes = func() (Boxes []Box) {
	for _, d := range Dots {
		if d.X() < BoardSize-1 && d.Y() < BoardSize-1 {
			Boxes = append(Boxes, Box(d))
		}
	}
	return Boxes
}()

// NewBoard 创建一个新的棋盘
func NewBoard(board Board) Board {
	newBoard := make(Board)
	for e := range board {
		newBoard[e] = struct{}{}
	}
	return newBoard
}

// EdgesCountInBox 计算方块中的边数
func (b Board) EdgesCountInBox(box Box) (count int) {
	boxEdges := box.Edges()
	for _, e := range boxEdges {
		if _, c := b[e]; c {
			count++
		}
	}
	return
}

// ObtainsScore 检查添加一条边后是否获得得分
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

// ObtainsBoxes 检查添加一条边后获得的方块
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
