package cache

import (
	"MTools/cache/lru"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var errNotFound = errors.New("not found")

func TestMaxMemoryCache_Set(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string
		val []byte

		wantKeys []string
		wantErr  error
		wantUsed int64
	}{
		{
			// 不触发淘汰
			name: "not exist",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{data: map[string][]byte{}})
				return res
			},
			key:      "key1",
			val:      []byte("hello"),
			wantKeys: []string{"key1"},
			wantUsed: 5,
		},
		{
			// 原本就有，覆盖导致 used 增加
			name: "override-incr",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello"),
					},
				})
				A := lru.Constructor()
				A.Put("key1")
				res.keys = A
				res.used = 5
				return res
			},
			key:      "key1",
			val:      []byte("hello,world"),
			wantKeys: []string{"key1"},
			wantUsed: 11,
		},
		{
			// 执行淘汰，一次
			name: "delete",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello, key1"),
						"key2": []byte("hello, key2"),
						"key3": []byte("hello, key3"),
					},
				})
				A1 := lru.Constructor()
				A1.Put("key1")
				A1.Put("key2")
				A1.Put("key3")
				res.keys = A1
				res.used = 33
				return res
			},
			key:      "key4",
			val:      []byte("hello, key4"),
			wantKeys: []string{"key4", "key2", "key3"},
			wantUsed: 33,
		},
		{
			// 执行淘汰，多次
			name: "delete-multi",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello, key1"),
						"key2": []byte("hello, key2"),
						"key3": []byte("hello, key3"),
					},
				})
				A2 := lru.Constructor()
				A2.Put("key1")
				A2.Put("key2")
				A2.Put("key3")
				res.keys = A2
				res.used = 33
				return res
			},
			key:      "key4",
			val:      []byte("hello, key4,hello, key4"),
			wantKeys: []string{"key4", "key3"},
			wantUsed: 34,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			err := cache.Set(context.Background(), tc.key, tc.val, time.Minute)
			assert.Equal(t, tc.wantKeys, cache.keys.AllKeys())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUsed, cache.used)
		})
	}
}

type mockCache struct {
	Cache
	fn   func(key string, val []byte)
	data map[string][]byte
}

func (m *mockCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.data[key] = val
	return nil
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		return val, nil
	}
	return nil, errNotFound
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	val, ok := m.data[key]
	if ok {
		m.fn(key, val)
	}
	return nil
}

func (m *mockCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		m.fn(key, val)
		return val, nil
	}
	return nil, errNotFound
}

func (m *mockCache) OnEvicted(fn func(key string, val []byte)) {
	m.fn = fn
}
