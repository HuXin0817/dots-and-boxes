package main

import (
	"log"
	"serve/serveclient"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

var ServeClient = NewServeClient()

func NewServeClient() serveclient.Serve {
	for range 60 {
		c, err := zrpc.NewClient(zrpc.RpcClientConf{Endpoints: ServerAddresses}, zrpc.WithNonBlock())
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}

		return serveclient.NewServe(c)
	}

	panic("Can not create new ServeClient")
}
