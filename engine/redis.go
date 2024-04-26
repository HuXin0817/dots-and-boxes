package main

import (
	"flag"

	"github.com/HuXin0817/dots-and-boxes/pkg/env"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	RedisConfigFile = flag.String("f", "develop.yaml", "redis config file path")
	rdbConf         redis.RedisConf
	RedisClient     *redis.Redis
)

func initConfig() {
	flag.Parse()
	conf.MustLoad(*RedisConfigFile, &rdbConf)
	if rdbConf.Pass == "" {
		rdbConf.Pass = env.RedisPassWord
	}

	RedisClient = redis.MustNewRedis(rdbConf)
}
