package ui

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	DotDistance = float32(75)
	DotWidth    = float32(15)
	DotMargin   = float32(50)
)

func getPosition(p int) float32 {
	return DotMargin + float32(p)*DotDistance
}

func getDotPosition(d chess.Dot) (float32, float32) {
	return getPosition(d.X()), getPosition(d.Y())
}

func NewDot(d chess.Dot) (newDotUI *canvas.Circle) {
	newDotUI = canvas.NewCircle(color.White)
	newDotUI.Resize(fyne.NewSize(DotWidth, DotWidth))
	newDotUI.Move(fyne.NewPos(getDotPosition(d)))
	return
}
