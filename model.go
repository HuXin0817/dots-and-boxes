package main

import (
	"fmt"
	"sync/atomic"

	"fyne.io/fyne/v2"
)

type Turn int

const (
	Player1Turn Turn = 1
	Player2Turn      = -Player1Turn
)

func (t *Turn) ToString() string {
	if *t == Player1Turn {
		return "Player1"
	} else {
		return "Player2"
	}
}

func (t *Turn) Change() { *t = -*t }

type Dot int

func NewDot(x, y int) Dot { return Dot(x*BoardSize + y) }

func (d Dot) X() int { return int(d) / BoardSize }

func (d Dot) Y() int { return int(d) % BoardSize }

func (d Dot) ToString() string { return fmt.Sprintf("(%d, %d)", d.X(), d.Y()) }

type Edge int

func NewEdge(Dot1, Dot2 Dot) Edge { return Edge(Dot1*BoardSizePower + Dot2) }

func (e Edge) Dot1() Dot { return Dot(e) / BoardSizePower }

func (e Edge) Dot2() Dot { return Dot(e) % BoardSizePower }

func (e Edge) ToString() string { return e.Dot1().ToString() + " => " + e.Dot2().ToString() }

func (e Edge) NearBoxes() []Box { return EdgeNearBoxes[e] }

type Box int

func (b Box) Edges() []Edge { return BoxEdges[b] }

type Board map[Edge]struct{}

func NewBoard(board Board) Board {
	b := make(Board, len(board))
	for e := range board {
		b[e] = struct{}{}
	}
	return b
}

func SendMessage(format string, a ...any) {
	App.SendNotification(&fyne.Notification{
		Title:   "Dots-And-Boxes",
		Content: fmt.Sprintf(format, a...),
	})
}

type Option struct {
	name  string
	value atomic.Bool
}

func NewOption(name string, value bool) *Option {
	op := &Option{name: name}
	op.value.Store(value)
	return op
}

func (op *Option) Value() bool { return op.value.Load() }

func (op *Option) On() {
	if !op.value.Load() {
		op.Change()
	}
}

func (op *Option) Off() {
	if op.value.Load() {
		op.Change()
	}
}

func (op *Option) Change() {
	if op.value.Load() {
		SendMessage(op.name + " OFF")
	} else {
		SendMessage(op.name + " ON")
	}
	op.value.Store(!op.value.Load())
}
