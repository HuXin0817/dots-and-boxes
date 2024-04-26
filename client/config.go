package main

import (
	"flag"
	"strconv"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/model"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/music"
)

var (
	AI1Conf       = flag.String("AI1", "ON", "AI1")
	AI2Conf       = flag.String("AI2", "ON", "AI2")
	BoardSizeConf = flag.String("BoardSize", "6", "BoardSize")
	MusicConf     = flag.String("Music", "On", "Music")

	BoardSize int
	AI1       model.Config
	AI2       model.Config
	GameUid   = message.NewGameUid()

	ServerAddresses = []string{
		"127.0.0.1:8000",
		// "127.0.0.1:8001",
		// "127.0.0.1:8002",
	}

	WithBGM = false
)

func initConfig() {
	flag.Parse()
	AI1 = model.NewConfig(*AI1Conf)
	AI2 = model.NewConfig(*AI2Conf)
	music.Open = model.NewConfig(*MusicConf)

	var err error
	BoardSize, err = strconv.Atoi(*BoardSizeConf)
	if err != nil {
		panic(err)
	}
}
