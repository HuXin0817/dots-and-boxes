package main

import (
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func GetFreeTopic(RedisClient *redis.Redis) (topic message.RedisPartition, err error) {
	for {
		for _, t := range message.RedisPartitions {
			value, err := RedisClient.Get(t.OwnerKey())
			if err != nil {
				return -1, err
			}

			length, err := RedisClient.Llen(t.ListKey())
			if err != nil {
				return -1, err
			}

			if value == ""  && length >0{
				if err = RedisClient.Setex(t.OwnerKey(), string(message.NewTimeStamp(time.Now())), 600); err != nil {
					return -1, err
				}
				return t, nil
			}

			time.Sleep(time.Second)
		}
	}
}
