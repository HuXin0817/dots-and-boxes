package model

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RedisLock struct {
	*redis.RedisLock
}

func NewLock(rds *redis.Redis, LockName string) *RedisLock {
	return &RedisLock{
		RedisLock: redis.NewRedisLock(rds, LockName),
	}
}

func (l *RedisLock) Do(f func() error) error {
	if err := l.Lock(); err != nil {
		return err
	}

	if err := f(); err != nil {
		return err
	}

	if err := l.UnLock(); err != nil {
		return err
	}

	return nil
}

func (l *RedisLock) Lock() error {
	acquire, err := l.Acquire()
	if err != nil {
		return err
	}

	if !acquire {
		time.Sleep(time.Second / 5)
		return l.Lock()
	}

	return nil
}

func (l *RedisLock) UnLock() error {
	release, err := l.Release()
	if err != nil {
		return err
	}

	if !release {
		time.Sleep(time.Second / 5)
		return l.UnLock()
	}

	return nil
}
