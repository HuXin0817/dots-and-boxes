package pprof

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func run() {
	router := gin.Default()
	pprof.Register(router)
	addr := fmt.Sprintf("localhost:%d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(0xffff))
	if err := router.Run(addr); err != nil {
		run()
	}
	time.Sleep(time.Second)
}

func init() {
	go run()
}
