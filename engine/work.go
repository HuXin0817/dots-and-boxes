package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/HuXin0817/dots-and-boxes/pkg/assess"
	"github.com/HuXin0817/dots-and-boxes/pkg/models/message"
)

func RollBack(t, m string) {
	for range 20 {
		if _, err := RedisClient.Lpush(t, m); err != nil {
			time.Sleep(time.Second / 2)
			RollBack(t, m)
		} else {
			return
		}
	}
}

func OnceIntervalWorking(NowTopic message.RedisPartition) (err error) {
	log.Printf("Start Working At Patition: %d\n", NowTopic)

	defer func() {
		if _, err = RedisClient.Del(NowTopic.OwnerKey()); err != nil {
			log.Panicf("Expire Error: %s\n", err)
		}
	}()

	for {
		if err = RedisClient.Expire(NowTopic.OwnerKey(), OnceWorkingTime); err != nil {
			return err
		}

		l, err := RedisClient.Llen(NowTopic.ListKey())
		if err != nil {
			return err
		}

		if l == 0 {
			return nil
		}

		m, err := RedisClient.Rpop(NowTopic.ListKey())
		if err != nil {
			return err
		}

		if m == "" {
			continue
		}

		fmt.Printf("=> %s\n", m)
		mess, err := message.NewMovingInformationMessage(m)
		if err != nil {
			return err
		}

		movingHasBeenAssessedKey := message.MovingHasBeenAssessedKey{
			GameUid: mess.GameUid,
			Step:    mess.StepCount,
			Edge:    mess.MoveEdge,
		}

		b, err := RedisClient.Get(movingHasBeenAssessedKey.String())
		if err != nil {
			return err
		}

		if b != "" {
			continue
		}

		nowStep, err := RedisClient.Get(string(mess.GameUid))
		if err != nil {
			RollBack(NowTopic.ListKey(), m)
			return err
		}

		if nowStep == "" {
			continue
		}

		if step, _ := strconv.Atoi(nowStep); step != mess.StepCount {
			continue
		}

		assessMessageKey := message.AssessMessageKey{
			GameUid: mess.GameUid,
			Step:    mess.StepCount,
		}

		move := assess.Move{
			Board: mess.Board,
			Edge:  mess.MoveEdge,
		}

		assessMessageValue := message.AssessMessageValue{
			Edge:  mess.MoveEdge,
			Score: assess.Assess(move),
		}

		Pusher.AddMessages(AssessMessage{
			AssessMessageKey:   assessMessageKey,
			AssessMessageValue: assessMessageValue,
			RollBackFunc: func() {
				RollBack(NowTopic.ListKey(), m)
			},
		})

		if err = RedisClient.Expire(assessMessageKey.String(), SetExpireTime); err != nil {
			return err
		}

		if err = RedisClient.Setex(movingHasBeenAssessedKey.String(), fmt.Sprint(assessMessageValue.Score), 240); err != nil {
			return err
		}
	}
}
