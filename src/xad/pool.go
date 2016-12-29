package xad

import (
	"time"

	"common"
	"xutil/pool"
)

var (
	// user getter pool
	execPool *pool.ObjectPool
)

func init() {
	if execPool == nil {
		execPool = pool.NewBufferPoolWithSize(10240,
			func() interface{} {
				return new(Executor)
			})
	}

	// 后台不断打印对象池的使用信息
	go PoolStatsDump()
}

func PoolStatsDump() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if execPool != nil {
				common.Infof("Object Pool Executor stats:%s", execPool.Json())
			}
		}
	}
	common.Warningf("PoolStatsDump quit goroutine")
}

func PickExecutor() *Executor {
	return execPool.Get().(*Executor)
}

func PutExecutor(e *Executor) {
	execPool.Put(e)
}
