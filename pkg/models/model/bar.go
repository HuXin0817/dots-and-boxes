package model

import (
	"github.com/logrusorgru/aurora"
	"github.com/schollz/progressbar/v3"
)

type Bar progressbar.ProgressBar

func NewBar(len int, description string) *Bar {
	return (*Bar)(progressbar.NewOptions(len,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        aurora.Yellow("█").String(),
			SaucerHead:    aurora.Yellow("█").String(),
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	))
}

func (b *Bar) Add(i int) {
	(*progressbar.ProgressBar)(b).Add(i)
}

func (b *Bar) Goto(i int) {
	(*progressbar.ProgressBar)(b).Set(i)
}

func (b *Bar) Close() {
	(*progressbar.ProgressBar)(b).Finish()
	(*progressbar.ProgressBar)(b).Close()
}
