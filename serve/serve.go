package main

import (
	"flag"
	"fmt"

	"serve/internal/config"
	"serve/internal/server"
	"serve/internal/svc"
	"serve/serve"

	_ "github.com/HuXin0817/dots-and-boxes/pkg/pprof"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configFile = flag.String("f", "etc/serve.yaml", "the config file")
	serveAddr  = flag.String("h", "0.0.0.0:8000", "the serve address")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	c.ListenOn = *serveAddr

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		serve.RegisterServeServer(grpcServer, server.NewServeServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
