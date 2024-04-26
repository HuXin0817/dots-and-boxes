package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	picture "github.com/HuXin0817/dots-and-boxes/png"
)

func AdjustImageTransparency(imgBytes []byte, factor float64) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			originalColor.A = uint8(float64(originalColor.A) * factor)
			dst.SetNRGBA(x, y, originalColor)
		}
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, dst); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var (
	imageData, _   = AdjustImageTransparency(picture.Gopher_png, 0.4)
	imgReader      = bytes.NewReader(imageData)
	imageObj, _, _ = image.Decode(imgReader)
)

func NewImage(x, y int) *canvas.Image {
	fyneImage := canvas.NewImageFromImage(imageObj)
	fyneImage.FillMode = canvas.ImageFillStretch
	fyneImage.Move(fyne.NewPos(getPosition(x)+DotWidth+5, getPosition(y)+DotWidth+5))
	fyneImage.Resize(fyne.NewSize(DotDistance-DotWidth-10, DotDistance-DotWidth-10))
	return fyneImage
}
