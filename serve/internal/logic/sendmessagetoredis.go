package logic

import (
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/chess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/pusher"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func GetMessages(GameUid message.GameUid, StepCount int, b chess.Board, edges ...chess.Edge) (messages []string) {
	for _, e := range edges {
		m := &message.MovingInformationMessage{
			TimeStamp: message.NewTimeStamp(time.Now()),
			GameUid:   GameUid,
			StepCount: StepCount,
			Board:     b,
			MoveEdge:  e,
		}
		messages = append(messages, m.String())
	}

	return
}

func GetTopicMessageList(RedisClient *redis.Redis, messages []string) (topicMessageList map[message.RedisPartition][]string, err error) {
	topicListLen := make(map[message.RedisPartition]int)
	for _, t := range message.RedisPartitions {
		topicListLen[t], err = RedisClient.Llen(t.ListKey())
		if err != nil {
			return nil, err
		}
	}

	topicMessageList = make(map[message.RedisPartition][]string)
	for i := 0; i < len(messages); i++ {
		minTopic := message.RedisPartition(-1)
		minLen := 0
		for t, length := range topicListLen {
			if minTopic == -1 || length < minLen {
				minLen = length
				minTopic = t
			}
		}

		topicListLen[minTopic]++
		topicMessageList[minTopic] = append(topicMessageList[minTopic], messages[i])
	}

	return topicMessageList, nil
}

func SendMessageToRedisLists(RedisClient *redis.Redis, PartitionPusher map[message.RedisPartition]*pusher.Pusher[string], GameUid message.GameUid, StepCount int, b chess.Board, edges ...chess.Edge) (err error) {
	messages := GetMessages(GameUid, StepCount, b, edges...)

	topicMessageList, err := GetTopicMessageList(RedisClient, messages)
	if err != nil {
		return err
	}

	for part, mess := range topicMessageList {
		PartitionPusher[part].AddMessages(mess...)
	}

	return nil
}
