package ui

import (
	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	FilledColor           = color.RGBA{R: 128, G: 128, B: 128, A: 128}
	Player1HighLightColor = color.NRGBA{R: 30, G: 30, B: 255, A: 128}
	Player2HighLightColor = color.NRGBA{R: 255, G: 30, B: 30, A: 128}
)

func NewEdge(e chess.Edge) *canvas.Line {
	x1 := getPosition(e.Dot1().X()) + DotWidth/2
	y1 := getPosition(e.Dot1().Y()) + DotWidth/2
	x2 := getPosition(e.Dot2().X()) + DotWidth/2
	y2 := getPosition(e.Dot2().Y()) + DotWidth/2

	newEdge := canvas.NewLine(FilledColor)
	newEdge.Position1 = fyne.NewPos(x1, y1)
	newEdge.Position2 = fyne.NewPos(x2, y2)
	newEdge.StrokeWidth = DotWidth
	return newEdge
}
