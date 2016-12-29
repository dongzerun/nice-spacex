package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"common"
	"config"
	"defined"
)

var (
	GlobalRedisCache *RedisWrap

	GlobalShowCache *RedisWrap

	GlobalSocialCache *RedisWrap
)

type RedisWrap struct {
	pools []*common.RedisPool
	retry int

	// protecting follow
	lock        sync.Mutex
	RoundRobinQ []int
	Weights     []int
	LastPoolIdx int
}

func init() {
	defined.RegisterOnRun("cache", func() {
		common.Info("Server On Run cache.InitCache")
		InitCache()
	})
}

func InitCache() {
	if GlobalRedisCache == nil {
		GlobalRedisCache = NewRedisWrap(config.GlobalConfig.CacheConfig)
	}

	if GlobalShowCache == nil {
		GlobalShowCache = NewRedisWrap(config.GlobalConfig.ShowCacheConfig)
	}

	if GlobalSocialCache == nil {
		GlobalSocialCache = NewRedisWrap(config.GlobalConfig.SocialConfig)
	}
}

func NewRedisWrap(url map[string]config.Redisconfig) *RedisWrap {
	c := &RedisWrap{
		pools: make([]*common.RedisPool, 0, 10),
		retry: 1,
	}

	for i, _ := range url {
		p, err := common.NewRedisPool(
			url[i].Host,
			url[i].Port,
			url[i].MaxIdle,
			url[i].IdleTimeout,
			url[i].ConnectTimeout,
			url[i].ReadTimeout,
			url[i].WriteTimeout,
			url[i].PoolSize)
		if err != nil {
			common.Warningf("%s %d connected failed %s", url[i].Host, url[i].Port, err.Error())
			continue
		}
		c.pools = append(c.pools, p)
	}

	if len(c.pools) == 0 {
		panic("redis Wrap have no pool, quit main...")
	}

	c.Weights = make([]int, len(c.pools))

	for i := 0; i < len(c.Weights); i++ {
		c.Weights[i] = 1
	}
	c.InitBalancer()
	return c
}

func (c *RedisWrap) ScoreInt64(key string, member interface{}) (int64, error) {
	var (
		score int64
		err   error
		p     *common.RedisPool
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		score, err = p.Int64("ZSCORE", key, member)
		if err == nil {
			return score, nil
		}
		if err.Error() != "redigo: nil returned" {
			return 0, err
		}
	}
	return -1, err
}

func (c *RedisWrap) GetInt64(key string) (int64, error) {
	var (
		cnt int64
		err error
		p   *common.RedisPool
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		cnt, err = p.Int64("GET", key)
		if err == nil {
			return cnt, nil
		}
		if err.Error() != "redigo: nil returned" {
			return 0, err
		}
	}
	return -1, err
}

func (c *RedisWrap) GetByte(key string) ([]byte, error) {
	var (
		data []byte
		err  error
		p    *common.RedisPool
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		data, err = p.Bytes("GET", key)
		if err == nil {
			return data, nil
		}
		if err.Error() != "redigo: nil returned" {
			return nil, err
		}
	}
	return nil, err
}

func (c *RedisWrap) GetInt(key string) (int, error) {
	var (
		cnt int
		err error
		p   *common.RedisPool
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		cnt, err = p.Int("GET", key)
		if err == nil {
			return cnt, nil
		}
		if err.Error() != "redigo: nil returned" {
			return 0, err
		}
	}
	return -1, err
}

func (c *RedisWrap) Exists(key string) (bool, error) {
	var (
		err    error
		exists bool
		p      *common.RedisPool
	)

	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		exists, err = p.Bool("EXISTS", key)
		if err == nil {
			return exists, nil
		}
	}
	return false, err
}

func (c *RedisWrap) Incr(key string) (int64, error) {
	var (
		p   *common.RedisPool
		v   int64
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		v, err = p.Int64("INCR", key)
		if err == nil {
			return v, nil
		}
	}
	return -1, err
}

func (c *RedisWrap) IncrBy(key string, delta int64) (int64, error) {
	var (
		p   *common.RedisPool
		v   int64
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		v, err = p.Int64("INCRBY", key, delta)
		if err == nil {
			return v, nil
		}
	}
	return -1, err
}

func (c *RedisWrap) Del(key string) error {
	var (
		p   *common.RedisPool
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		_, err = p.Do("DEL", key)
		return err
	}
	return err
}

func (c *RedisWrap) SetWithTTL(key string, count int64, ttl int64) error {
	var (
		p   *common.RedisPool
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		p.Bool("SET", key, count)
		_, err = p.Int("EXPIRE", key, ttl)
		if err == nil {
			return nil
		}
	}
	return err
}

// SETEX cache_user_id 60 10086
func (c *RedisWrap) SetByteWithEx(key string, ctx []byte, ttl int64) error {
	var (
		p   *common.RedisPool
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		_, err = p.Do("SETEX", key, ttl, ctx)
		return err
	}
	return err
}

func (c *RedisWrap) SetTTL(key string, ttl int64) error {
	var (
		p   *common.RedisPool
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		_, err = p.Int("EXPIRE", key, ttl)
		if err == nil {
			return nil
		}
	}
	return err
}

func (c *RedisWrap) Set(key string, count int64) error {
	var (
		p   *common.RedisPool
		err error
	)
	for i := 0; i < c.retry; i++ {
		p, err = c.GetNextPool()
		if err != nil {
			continue
		}
		_, err = p.String("SET", key, count)
		if err == nil {
			return nil
		}
	}
	return err
}

func Gcd(ary []int) int {
	var i int
	min := ary[0]
	length := len(ary)
	for i = 0; i < length; i++ {
		if ary[i] < min {
			// 找到最小的 min weight权重值
			min = ary[i]
		}
	}

	// min = 8
	for {
		isCommon := true
		for i = 0; i < length; i++ {
			// 10 % 8 !=0成立，
			if ary[i]%min != 0 {
				isCommon = false
				break
			}
		}
		if isCommon {
			break
		}
		min--
		if min < 1 {
			break
		}
	}
	return min
}

func (c *RedisWrap) InitBalancer() {
	var sum int
	c.LastPoolIdx = 0
	gcd := Gcd(c.Weights)

	for _, weight := range c.Weights {
		sum += weight / gcd
	}

	c.RoundRobinQ = make([]int, 0, sum)
	for index, weight := range c.Weights {
		for j := 0; j < weight/gcd; j++ {
			c.RoundRobinQ = append(c.RoundRobinQ, index)
		}
	}

	if 1 < len(c.Weights) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < sum; i++ {
			x := r.Intn(sum)
			temp := c.RoundRobinQ[x]
			other := sum % (x + 1)
			c.RoundRobinQ[x] = c.RoundRobinQ[other]
			c.RoundRobinQ[other] = temp
		}
	}
}

func (c *RedisWrap) GetNextPool() (*common.RedisPool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var index int
	if len(c.RoundRobinQ) == 0 {
		return nil, fmt.Errorf("error no redis database")
	}
	if len(c.RoundRobinQ) == 1 {
		index = c.RoundRobinQ[0]
		return c.pools[index], nil
	}

	queueLen := len(c.RoundRobinQ)
	index = c.RoundRobinQ[c.LastPoolIdx]
	p := c.pools[index]
	c.LastPoolIdx++
	if queueLen <= c.LastPoolIdx {
		c.LastPoolIdx = 0
	}
	return p, nil
}
