package xuser

import (
	"os"
	"strings"
	"time"

	"common"
	cache "xutil/lru"
)

var (
	lru      *InnerLruCache
	Hostname string
)

func init() {
	lru = NewInnerLruCache(10240)

	host, err := os.Hostname()
	if err != nil {
		panic("get hostname failed " + err.Error())
	}
	Hostname = strings.Split(host, ".")[0]
}

type UserItem struct {
	Data []byte
}

// Value 必须实现 Size() int 方法
func (ui *UserItem) Size() int {
	return 1
}

type InnerLruCache struct {
	// 全局的 LRU Cache
	lru *cache.LRUCache
	// 用于统计 LRU 命中率的两个变量和管道
	hit  int64
	miss int64
	// 传递数据 0 表示miss,1 表示命中数据
	lruChan chan int
}

func NewInnerLruCache(capicty int64) *InnerLruCache {
	ilc := &InnerLruCache{
		lru:     cache.NewLRUCache(capicty),
		lruChan: make(chan int, 1024),
	}

	go ilc.loopStats()
	return ilc
}

func (ilc *InnerLruCache) Get(key string) ([]byte, bool) {
	v, ok := ilc.lru.Get(key)
	if ok {
		// 命中 hit +1
		ilc.lruChan <- 1
	} else {
		// 未命中 miss +1
		ilc.lruChan <- 0
		return nil, false
	}

	// 类型转换到 *UserItem
	if ui, ok := v.(*UserItem); ok {
		return ui.Data, true
	}
	return nil, false
}

// kv写到 LRU 中
func (ilc *InnerLruCache) Set(key string, v []byte) {
	ui := &UserItem{
		Data: v,
	}
	ilc.lru.Set(key, ui)
}

// 使Cache失效
func (ilc *InnerLruCache) Clear() {
	ilc.lru.Clear()
}

func (ilc *InnerLruCache) StatsJSON() string {
	return ilc.lru.StatsJSON()
}

func (ilc *InnerLruCache) loopStats() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	clear := time.NewTicker(3600 * time.Second)
	defer clear.Stop()

	for {
		select {
		// 接收统计数据
		case d := <-ilc.lruChan:
			if d == 0 {
				ilc.miss += 1
			} else if d == 1 {
				ilc.hit += 1
			}
		// 每分钟将统计数据打到 influxdb
		case <-ticker.C:
			var pct int
			if (ilc.miss + ilc.hit) == 0 {
				pct = 0
			} else {
				pct = int(ilc.hit * 100 / (ilc.miss + ilc.hit))
			}

			common.Info("InnerLruCache hit ", ilc.hit, " miss ", ilc.miss)
			common.Info("InnerLruCache JSON STATS ", ilc.StatsJSON())
			go pushMetric("hit", pct)
			ilc.miss = 0
			ilc.hit = 0
		// 每1小时失效一次内置 Cache
		// 由于二级 Redis Cache 是6小时
		case <-clear.C:
			ilc.Clear()
		}
	}
	common.Warning("InnerLruCache loopStats quit goroutine")
}

func pushMetric(name string, num int) {
	tags := map[string]string{
		"hostname": Hostname,
		"metric":   name,
	}
	data := map[string]interface{}{
		"pct": num,
	}

	err := common.PutMetric("lrucache", tags, data)
	if err != nil {
		common.Warning("statsLoop PutMetric err ", err.Error())
	}
}
