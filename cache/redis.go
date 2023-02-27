package cache

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"time"
)

type RedisCacheLock struct {
	Client redis.Cmdable
}

func (r RedisCacheLock) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := r.Client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("redis-lock: 抢锁失败")
	}
	return &Lock{
		client:     r.Client,
		key:        key,
		value:      val,
		expiration: expiration,
	}, nil
}

func (r RedisCacheLock) Lock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	return &Lock{}, nil
}

func (r RedisCacheLock) UnLock() (*Lock, error) {
	return &Lock{}, nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
	unlockChan chan struct{}
}
