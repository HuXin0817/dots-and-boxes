package music

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

func BeepPlay(mp3Bytes []byte) {
	if !Open {
		return
	}

	reader := nopCloser{bytes.NewReader(mp3Bytes)}

	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Fatalf("Error decoding MP3: %v", err)
	}

	defer func(streamer beep.StreamSeekCloser) {
		err := streamer.Close()
		if err != nil {
			log.Printf("Error closing stream: %v", err)
		}
	}(streamer)

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		log.Fatalf("Error initializing speaker: %v", err)
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
