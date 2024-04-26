package main

import (
	"context"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/music"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	initConfig()
	initWindow()
	if WithBGM {
		go music.Play()
	}

	go Exit()
	Window.Run()
}

func Exit() {
	time.Sleep(2 * time.Second)
	addEdgeLogic()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	<-sigChan

	r := GetPostGameInformationRequest()
	r.GameOver = true
	if _, err := ServeClient.PostGameInformation(context.Background(), GetPostGameInformationRequest()); err != nil {
		panic(err)
	}

	os.Exit(0)
}
