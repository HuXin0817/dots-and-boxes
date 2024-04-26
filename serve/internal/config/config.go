package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MongoConf struct {
		Url          string
		DataBaseName string
		PassWord     string `json:",optional"`
	}
}
