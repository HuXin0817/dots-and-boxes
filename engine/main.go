package main

import (
	"log"
	"time"

	_ "github.com/HuXin0817/dots-and-boxes/pkg/pprof"
)

const (
	OnceWorkingTime = 180                 // second
	SetExpireTime   = OnceWorkingTime * 3 // second
)

func main() {
	initConfig()
	Pusher.Start()

	for {
		NowTopic, err := GetFreeTopic(RedisClient)
		if err != nil {
			log.Fatalln(err)
		}

		if err = OnceIntervalWorking(NowTopic); err != nil {
			log.Fatalln(err)
		}

		time.Sleep(time.Second)
	}
}
