package cache

import (
	"MTools/lua"
	"MTools/mocks"
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTryLock(t *testing.T) {
	testcases := []struct {
		name     string
		key      string
		mocks    func(ctrl *gomock.Controller) redis.Cmdable
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "empty key",
			key:  "key1",
			mocks: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			wantLock: &Lock{
				client:     nil,
				key:        "key1",
				value:      "",
				expiration: time.Minute,
				unlockChan: nil,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := NewClient(tc.mocks(ctrl))
			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.Equal(t, tc.wantLock.expiration, l.expiration)
			// 赋予值了
			assert.NotEmpty(t, l.value)
		})
	}
}

//	func TestLock(t *testing.T) {
//		t.Parallel()
//		testCases := []struct {
//			name       string
//			mocks      func(ctrl *gomock.Controller) redis.Cmdable
//			key        string
//			value      string
//			expiration time.Duration
//			timeout    time.Duration
//			retry      RetryStrategy
//			wantLock   *Lock
//			wantErr    error
//		}{
//			{
//				name: "locked",
//				mocks: func(ctrl *gomock.Controller) redis.Cmdable {
//					cmdable := mocks.NewMockCmdable(ctrl)
//					res := redis.NewCmd(context.Background(), nil)
//					res.SetVal("OK")
//					cmdable.EXPECT().Eval(gomock.Any(), lua.LuaLock, []string{"locked-key"}, gomock.Any()).
//						Return(res)
//					return cmdable
//				},
//				key:        "lock_key1",
//				value:      "lock_value1",
//				expiration: time.Minute,
//				timeout:    time.Second * 3,
//				retry: &FixedIntervalRetryStrategy{
//					Interval: time.Second,
//					MaxCnt:   10,
//				},
//				wantLock: &Lock{
//					key:        "lock_key1",
//					expiration: time.Minute,
//				},
//			},
//		}
//
//		for _, tc := range testCases {
//			t.Run(tc.name, func(t *testing.T) {
//				ctrl := gomock.NewController(t)
//				defer ctrl.Finish()
//				client := NewClient(tc.mocks(ctrl))
//				lock, err := client.Lock(context.Background(), tc.key, tc.expiration, tc.timeout, tc.retry)
//				if err != nil {
//					return
//				}
//				assert.Equal(t, tc.wantLock.key, lock.key)
//				assert.Equal(t, tc.wantLock.expiration, lock.expiration)
//				assert.NotEmpty(t, lock.value)
//				assert.NotNil(t, lock.client)
//			})
//		}
//	}
func TestClient_Lock(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testCases := []struct {
		name string

		mock func() redis.Cmdable

		key        string
		expiration time.Duration
		retry      RetryStrategy
		timeout    time.Duration

		wantLock *Lock
		wantErr  string
	}{
		{
			name: "locked",
			mock: func() redis.Cmdable {
				cmdable := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background(), nil)
				redis.NewStringCmd(context.Background(), nil)
				res.SetVal("OK")
				cmdable.EXPECT().Eval(gomock.Any(), lua.LuaLock, []string{"locked-key1"}, gomock.Any()).
					Return(res)
				return cmdable
			},
			key:        "locked-key1",
			expiration: time.Minute,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			timeout: time.Second,
			wantLock: &Lock{
				key:        "locked-key1",
				expiration: time.Minute,
			},
		},
		{
			name: "retry and success",
			mock: func() redis.Cmdable {
				cmdable := mocks.NewMockCmdable(ctrl)
				first := redis.NewCmd(context.Background(), nil)
				first.SetVal("")
				cmdable.EXPECT().Eval(gomock.Any(), lua.LuaLock, []string{"retry-key"}, gomock.Any()).
					Times(2).Return(first)
				second := redis.NewCmd(context.Background(), nil)
				second.SetVal("OK")
				cmdable.EXPECT().Eval(gomock.Any(), lua.LuaLock, []string{"retry-key"}, gomock.Any()).
					Return(second)
				return cmdable
			},
			key:        "retry-key",
			expiration: time.Minute,
			retry:      &FixedIntervalRetryStrategy{Interval: time.Millisecond, MaxCnt: 3},
			timeout:    time.Second,
			wantLock: &Lock{
				key:        "retry-key",
				expiration: time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRedisCmd := tc.mock()
			client := NewClient(mockRedisCmd)
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			l, err := client.Lock(ctx, tc.key, tc.expiration, time.Second, tc.retry)
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, mockRedisCmd, l.client)
			assert.Equal(t, tc.key, l.key)
			assert.NotEmpty(t, l.value)
			assert.Equal(t, tc.expiration, l.expiration)

		})
	}
}
