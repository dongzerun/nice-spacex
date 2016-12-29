package cache

import (
	"testing"
	"time"

	. "xutil"
)

func TestRR(t *testing.T) {
	// defer func() {
	// 	err := recover()
	// 	if err != nil {
	// 		const size = 4096
	// 		buf := make([]byte, size)
	// 		buf = buf[:runtime.Stack(buf, false)]
	// 		log.Print("stack", string(buf))
	// 	}

	// }()

	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)

	p, err := c.GetNextPool()
	if err != nil {
		t.Errorf("c.GetNextPool must ok %s", err.Error())
	}

	if p == nil {
		t.Errorf("c.GetNextPool must get p != nil ")
	}

	p.Int("DEL", "testkey")

	exists, e := c.Exists("testkey")
	if e != nil {
		t.Errorf("c.Exists error ", e.Error())
	}

	if exists {
		t.Errorf("c.Exists testkey must false but get true ")
	}

	_, er := c.GetInt64("testkey")
	// log.Print(er.Error() == "redigo: nil returned")
	if er == nil {
		t.Errorf("c.Get  must get empty error ")
	}
	if er.Error() != "redigo: nil returned" {
		t.Errorf("c.Get error must be redigo: nil returned, buf get %s ", er.Error())
	}

	err = c.Set("testkey", 1000)

	if err != nil {
		t.Errorf("c.Set get error %s", err.Error())
	}

	cnt, err1 := c.GetInt64("testkey")
	if err1 != nil {
		t.Errorf("c.Get  error %s", err1.Error())
	}

	if cnt != 1000 {
		t.Errorf("c.Get expected 1000,but get %d", cnt)
	}

	_, err = c.Incr("testkey")
	if err != nil {
		t.Errorf("c.Incr get error %s", err.Error())
	}

	cnt, err1 = c.GetInt64("testkey")
	if err1 != nil {
		t.Errorf("c.Get  error %s", err1.Error())
	}

	if cnt != 1001 {
		t.Errorf("c.Get expected 1001,but get %d", cnt)
	}

	err = c.SetWithTTL("testkey", 1000, 2000)
	if err != nil {
		t.Errorf("c.SetWithTTL get error %s", err.Error())
	}

	p, err = c.GetNextPool()
	if err != nil {
		t.Errorf("c.GetNextPool must ok %s", err.Error())
	}

	if p == nil {
		t.Errorf("c.GetNextPool must get p != nil ")
	}

	p.Int("DEL", "testkey")
}

func Benchmark_GetInt64(b *testing.B) {
	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)
	for i := 0; i < b.N; i++ {
		c.GetInt64("testkey")
	}
}

func Benchmark_Exists(b *testing.B) {
	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)
	for i := 0; i < b.N; i++ {
		c.Exists("testkey")
	}
}

func Benchmark_Incr(b *testing.B) {
	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)
	for i := 0; i < b.N; i++ {
		c.Incr("testkey")
	}
}

func Benchmark_Set(b *testing.B) {
	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)
	for i := 0; i < b.N; i++ {
		c.Set("testkey", 1000)
	}
}

func Benchmark_SetWithTTL(b *testing.B) {
	cfg1 := Redisconfig{}
	cfg1.Host = "127.0.0.1"
	cfg1.Port = 6379
	cfg1.MaxIdle = 5
	cfg1.PoolSize = 10
	cfg1.IdleTimeout = 30 * time.Second
	cfg1.ReadTimeout = 3 * time.Second
	cfg1.WriteTimeout = 3 * time.Second
	cfg1.ConnectTimeout = 1 * time.Second

	cfg := make([]Redisconfig, 3)
	cfg[0] = cfg1
	cfg[1] = cfg1
	cfg[2] = cfg1

	c := NewRedisWrap(cfg)
	for i := 0; i < b.N; i++ {
		c.SetWithTTL("testkey", 1000, 2000)
	}
}
