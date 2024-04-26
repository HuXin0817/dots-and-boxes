package music

import (
	"bytes"
	"io"
	"log"

	"github.com/hajimehoshi/go-mp3"
	"github.com/timshannon/go-openal/openal"
)

func OpenAlPlay(mp3Bytes []byte) {
	if !Open {
		return
	}

	reader := bytes.NewReader(mp3Bytes)

	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		log.Fatalf("failed to create MP3 decoder: %v", err)
	}

	device := openal.OpenDevice("")
	if device == nil {
		log.Fatal("Failed to open a device.")
	}
	defer device.CloseDevice()

	context := device.CreateContext()
	if context == nil {
		log.Fatal("Failed to create a context.")
	}
	defer context.Destroy()

	context.Activate()

	source := openal.NewSource()
	buffer := openal.NewBuffer()
	defer source.Stop()

	pcmBytes, err := io.ReadAll(decoder)
	if err != nil {
		log.Fatalf("failed to read PCM data: %v", err)
	}

	buffer.SetData(openal.FormatStereo16, pcmBytes, int32(decoder.SampleRate()))
	source.SetBuffer(buffer)

	source.Play()
	for source.State() == openal.Playing {
	}

	log.Println("Playback finished.")
}
