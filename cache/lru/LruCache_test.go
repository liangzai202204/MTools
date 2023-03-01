package lru

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLruCache_Get(t *testing.T) {
	cache := Constructor()
	cache.Put("key1")
	cache.Put("key2")
	v1 := cache.Get("key1")
	assert.Equal(t, "key1", v1)
	delete(cache.Cache, "key1")
	v2 := cache.Get("key1")
	assert.Equal(t, "", v2)
}
