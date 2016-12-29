package common_test

import (
	. "common"
	"fmt"
	"testing"
)

func TestRedis(t *testing.T) {
	redisHandle, err := NewRedisPoolWithConfig("redis.cfg", "local_redis")
	if err != nil {
		return
	}
	key := fmt.Sprintf("key:%d", 1)
	redisHandle.Bool("SET", key, 100)
	fmt.Println(redisHandle.Expire(key, 1000))
	fmt.Println(redisHandle.Int("GET", key))
}
