package svc

import (
	"fmt"
	"serve/internal/config"

	"github.com/HuXin0817/dots-and-boxes/pkg/env"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/model"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/pusher"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config          config.Config
	RedisClient     *redis.Redis
	PartitionPusher map[message.RedisPartition]*pusher.Pusher[string]
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.Redis.Pass == "" {
		c.Redis.Pass = env.RedisPassWord
	}

	if c.MongoConf.PassWord == "" {
		c.MongoConf.PassWord = env.MongoPassWord
	}

	c.MongoConf.Url = fmt.Sprintf(c.MongoConf.Url, c.MongoConf.PassWord)

	partitionListPushLock := make(map[message.RedisPartition]*model.RedisLock)

	svcCtx := &ServiceContext{
		Config:          c,
		RedisClient:     redis.MustNewRedis(c.Redis.RedisConf),
		PartitionPusher: make(map[message.RedisPartition]*pusher.Pusher[string]),
	}

	for _, redisPartition := range message.RedisPartitions {
		partitionListPushLock[redisPartition] = model.NewLock(svcCtx.RedisClient, redisPartition.LockName())

		svcCtx.PartitionPusher[redisPartition] = pusher.NewPusher(pusher.WithPushLogic(func(pushMessages ...string) error {
			if len(pushMessages) == 0 {
				return nil
			}

			return partitionListPushLock[redisPartition].Do(func() (err error) {
				var messages []any
				for _, m := range pushMessages {
					messages = append(messages, m)
				}

				if _, err = svcCtx.RedisClient.Lpush(redisPartition.ListKey(), messages...); err != nil {
					return err
				}

				redisPartitionLength, err := svcCtx.RedisClient.Llen(redisPartition.ListKey())
				if err != nil {
					return err
				}

				if err = svcCtx.RedisClient.Expire(redisPartition.ListKey(), 120*redisPartitionLength); err != nil {
					return err
				}

				return nil
			})
		}))

		svcCtx.PartitionPusher[redisPartition].Start()
	}

	return svcCtx
}
