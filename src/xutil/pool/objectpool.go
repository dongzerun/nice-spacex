package pool

import (
	"fmt"
	"sync/atomic"
)

type ObjectPool struct {
	// 其实用 channel 模拟也会有锁开销，要视情况而定
	c   chan interface{}
	New func() interface{}

	// stats
	hit  int64
	miss int64
}

// 初始化传入 New 工厂方法
func NewBufferPool(fn func() interface{}) *ObjectPool {
	return &ObjectPool{
		c:   make(chan interface{}, 4096),
		New: fn,
	}
}

// 初始化传入 size大小和 New 工厂方法
func NewBufferPoolWithSize(size int, fn func() interface{}) *ObjectPool {
	return &ObjectPool{
		c:   make(chan interface{}, size),
		New: fn,
	}
}

// 从对象池中获取
func (op *ObjectPool) Get() (o interface{}) {
	select {
	case o = <-op.c:
		atomic.AddInt64(&op.hit, 1)
	default:
		// 池中为空，那么新建一个
		o = op.New()
		atomic.AddInt64(&op.miss, 1)
	}
	return
}

// 将对象放回池中
func (op *ObjectPool) Put(o interface{}) {
	select {
	// 扔回池中，或是直接丢弃，交给 GC 回收
	case op.c <- o:
	default:
	}

}

func (op *ObjectPool) Json() string {
	h := atomic.LoadInt64(&op.hit)
	m := atomic.LoadInt64(&op.miss)
	r := float64(h) * 100 / (float64(h) + float64(m))
	return fmt.Sprintf("{\"Hit\": %d, \"Miss\": %d, \"Hit_Rate\": %.3f%%}", h, m, r)
}
