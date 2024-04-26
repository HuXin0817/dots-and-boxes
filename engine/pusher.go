package main

import (
	"fmt"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/model"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/pusher"

	"time"
)

type AssessMessage struct {
	message.AssessMessageKey
	message.AssessMessageValue
	RollBackFunc func()
}

var MessagePushLockMap = make(map[string]*model.RedisLock)

var Pusher = pusher.NewPusher(pusher.WithPushInterval[AssessMessage](time.Second), pusher.WithPushLogic(func(assessMessage ...AssessMessage) error {
	for _, assessMessage := range assessMessage {
		keyStr := assessMessage.AssessMessageKey.String()

		if MessagePushLockMap[keyStr] == nil {
			MessagePushLockMap[keyStr] = model.NewLock(RedisClient, fmt.Sprintf("%s-Lock", keyStr))
			go func() {
				time.Sleep(time.Minute * 2)
				delete(MessagePushLockMap, keyStr)
			}()
		}

		err := MessagePushLockMap[keyStr].Do(func() error {
			if _, err := RedisClient.Sadd(keyStr, assessMessage.AssessMessageValue.String()); err != nil {
				return err
			}

			if err := RedisClient.Expire(keyStr, SetExpireTime); err != nil {
				assessMessage.RollBackFunc()
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}))
