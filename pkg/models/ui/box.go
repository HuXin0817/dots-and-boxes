package ui

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	Player1Color = color.NRGBA{R: 30, G: 30, B: 128, A: 128}
	Player2Color = color.NRGBA{R: 128, G: 30, B: 30, A: 128}
)

var BoxSize = DotDistance - DotWidth

func NewBox(s chess.Box) *canvas.Rectangle {
	d := chess.Dot(s)
	x := getPosition(d.X()) + DotWidth
	y := getPosition(d.Y()) + DotWidth

	r := canvas.NewRectangle(color.Black)
	r.Move(fyne.NewPos(x, y))
	r.Resize(fyne.NewSize(BoxSize, BoxSize))
	return r
}
