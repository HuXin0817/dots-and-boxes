package music

import (
	"github.com/HuXin0817/dots-and-boxes/bgm"
	"time"
)

func Play() {
	for {
		for _, m := range bgm.MusicList {
			OpenAlPlay(m)
			time.Sleep(time.Second >> 1)
		}
	}
}
