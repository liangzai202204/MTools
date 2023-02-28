package cache

import (
	"MTools/lua"
	"context"
	"errors"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"time"
)

type RedisCacheLock struct {
	Client redis.Cmdable
}

func NewClient(client redis.Cmdable) *RedisCacheLock {
	return &RedisCacheLock{Client: client}
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

func (r RedisCacheLock) Lock(ctx context.Context, key string, expiration time.Duration, to time.Duration, retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		ctxTime, cancel := context.WithTimeout(ctx, to)
		cmd, err := r.Client.Eval(
			ctxTime,
			lua.LuaLock,
			[]string{key},
			val,
			expiration).Result()
		cancel()
		if err != nil {
			return nil, err
		}
		if cmd == "OK" {
			return &Lock{
				client:     r.Client,
				key:        key,
				value:      val,
				expiration: expiration,
				unlockChan: make(chan struct{}, 1),
			}, nil
		}
		// start retry config
		retryTime, next := retry.Next()
		if !next {
			return nil, errors.New("重试枪锁失败")
		}
		if timer == nil {
			timer = time.NewTimer(retryTime)
		} else {
			timer.Reset(retryTime)
		}
		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
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

// Refresh 执行一段lua脚本，判断key、val是不是自己的
// 如果是则刷新，否则return 0
func (l *Lock) Refresh(ctx context.Context) error {
	cmd, err := l.client.Eval(ctx,
		lua.LuaRefresh,
		[]string{l.key},
		l.value,
		l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}
	if cmd != 1 {
		return errors.New("redis-lock: 没有锁")
	}
	return nil
}

// AutoRefresh interval是自动刷新时间，timeout是超时时间，自动刷新会进入循环
// 利用time.Ticker定时器触发刷新机制（调用refresh手动刷新） 如何处理调用refresh手动刷新返回的error？
// 1、超时则重新触发刷新机制（timeoutChan <- struct{}{}），循环
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-timeoutChan:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.unlockChan:
			return nil
		}
	}
}
