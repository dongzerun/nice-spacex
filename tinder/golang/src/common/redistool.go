package common

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"time"
)

var RedisPoolTimeOutError error

func init() {
	rand.Seed(time.Now().UnixNano())
	RedisPoolTimeOutError = errors.New("redis pool get conn timeout ")
}

type RedisPool struct {
	pools []*redis.Pool
	// pool 本身就实现了连接池，不需要channel去控制
	// channel chan bool
}

func NewRedisPoolWithConfig(conf string, db string) (*RedisPool, error) {
	config, err := NewConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("初始化redis配置出错")
	}
	host := config.MustValue(db, "host")
	port := config.MustInt(db, "port", 6379)
	maxIdle := config.MustInt(db, "maxIdle", 5)
	idleTimeout := time.Duration(config.MustInt(db, "idleTimeout", 3)) * time.Second
	connectTimeout := time.Duration(config.MustInt(db, "connectTimeout", 3)) * time.Second
	readTimeout := time.Duration(config.MustInt(db, "readTimeout", 3)) * time.Second
	writeTimeout := time.Duration(config.MustInt(db, "writeTimeout", 3)) * time.Second
	poolSize := config.MustInt(db, "poolSize", 10)

	pools, err := NewRedisPool(host, port, maxIdle, idleTimeout, connectTimeout, readTimeout, writeTimeout, poolSize)
	return pools, err
}

func NewRedisPool(host string, port int, maxIdle int, idleTimeout, connectTimeout, readTimeout, writeTimeout time.Duration, poolSize int) (*RedisPool, error) {
	var (
		url                  = fmt.Sprintf("%s:%d", host, port)
		redisPool *RedisPool = new(RedisPool)
		rpool     *redis.Pool
		rconn     redis.Conn
		err       error
	)
	if poolSize <= 0 {
		return nil, fmt.Errorf("Pool size 为空")
	}
	loopSize := poolSize
	for {
		if loopSize <= 0 {
			break
		}
		loopSize--
		rpool = newPool(url, maxIdle, idleTimeout, connectTimeout, readTimeout, writeTimeout)
		rconn = rpool.Get()

		_, err = rconn.Do("EXISTS", rand.Int63n(1000))
		rconn.Close()
		if err != nil {
			fmt.Println("连接redis出错")
			continue
		}
		redisPool.pools = append(redisPool.pools, rpool)
	}
	if len(redisPool.pools) == 0 {
		return nil, fmt.Errorf("创建的redis pool 为空")
	}
	// redisPool.channel = make(chan bool, poolSize)
	return redisPool, nil
}

// 单个redis连接池
func newPool(server string, maxIdle int, idleTimeout, connectTimeout, readTimeout, writeTimeout time.Duration) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			var (
				conn redis.Conn
				err  error
			)
			if conn, err = redis.DialTimeout("tcp", server, connectTimeout, readTimeout, writeTimeout); err != nil {
				return nil, err
			}
			if _, err = conn.Do("EXISTS", rand.Int63n(1000)); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, err
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			t = t.Add(idleTimeout)
			if time.Now().Before(t) {
				return nil
			}
			var err error
			_, err = conn.Do("EXISTS", rand.Int63n(1000))
			return err
		},
		Wait: true,
	}
}

func (this *RedisPool) GetPool() *redis.Pool {
	return this.pools[rand.Intn(len(this.pools))]
}

func (this *RedisPool) Release() {
	// <-this.channel
}

func (this *RedisPool) Expire(args ...interface{}) (bool, error) {
	pool := this.GetPool()
	if pool == nil {
		return false, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Bool(conn.Do("EXPIRE", args...))
}

func (this *RedisPool) Bool(commandName string, args ...interface{}) (bool, error) {
	pool := this.GetPool()
	if pool == nil {
		return false, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Bool(conn.Do(commandName, args...))
}

func (this *RedisPool) Bytes(commandName string, args ...interface{}) ([]byte, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Bytes(conn.Do(commandName, args...))
}

func (this *RedisPool) Float64(commandName string, args ...interface{}) (float64, error) {
	pool := this.GetPool()
	if pool == nil {
		return 0, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Float64(conn.Do(commandName, args...))
}

func (this *RedisPool) Int(commandName string, args ...interface{}) (int, error) {
	pool := this.GetPool()
	if pool == nil {
		return 0, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Int(conn.Do(commandName, args...))
}

func (this *RedisPool) Int64(commandName string, args ...interface{}) (int64, error) {
	pool := this.GetPool()
	if pool == nil {
		return 0, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Int64(conn.Do(commandName, args...))
}

func (this *RedisPool) Ints(commandName string, args ...interface{}) ([]int, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Ints(conn.Do(commandName, args...))
}

func (this *RedisPool) MultiBulk(commandName string, args ...interface{}) ([]interface{}, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.MultiBulk(conn.Do(commandName, args...))
}

func (this *RedisPool) String(commandName string, args ...interface{}) (string, error) {
	pool := this.GetPool()
	if pool == nil {
		return "", RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.String(conn.Do(commandName, args...))
}

func (this *RedisPool) Strings(commandName string, args ...interface{}) ([]string, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.Strings(conn.Do(commandName, args...))
}

func (this *RedisPool) StringMap(commandName string, args ...interface{}) (map[string]string, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return redis.StringMap(conn.Do(commandName, args...))
}

func (this *RedisPool) Do(commandName string, args ...interface{}) (interface{}, error) {
	pool := this.GetPool()
	if pool == nil {
		return nil, RedisPoolTimeOutError
	}

	conn := pool.Get()
	defer func() {
		conn.Close()
		this.Release()
	}()
	return conn.Do(commandName, args...)
}
