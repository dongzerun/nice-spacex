package limit

import (
	"time"

	"common"
	"xutil/misc"
)

// 针对exposure曝光设计的粒度控制服务
// 暂时不做成通用的, 将一天划分成 2 * 60 * 24 = 2880 个时间片
type LimitServer struct {
	// 起始时间点
	start int64
	// 曝光比例，每bucket num 个时间片里，只有valib bucket个时间片曝光
	// bucketNum 一定大于等于bucketValid
	bucketNum   int
	bucketValid int
	// 当前limitserver针对的tag做为唯一标记存在
	tag string
	// 做一个标记桶，来确定是否展示，N次内随机, 比较浪费内存，但是查找性能快
	buckets []bool
}

func (ls *LimitServer) Exposure() bool {
	bucket := (time.Now().Unix() - ls.start) / 30
	return ls.buckets[bucket%2880]
}

// 一定要保证 num, valid 均大于0，并且  num >= valid
func NewLimitServer(start int64, num int, valid int, tag string) *LimitServer {
	ts := &LimitServer{
		start:       start,
		bucketNum:   num,
		bucketValid: valid,
		buckets:     make([]bool, 2880),
	}

	for i := 0; i < 2880/num; i++ {
		rand := misc.RandomInt(num)
		for n := 0; n < valid; n++ {
			idx := (rand + n) % num
			ts.buckets[i*num+idx] = true
		}
	}
	common.Infof("tag:%s, start:%d, num:%d, valid:%d,bucket:%v", ts.tag, ts.start, ts.bucketNum, ts.bucketValid, ts.buckets)

	return ts
}
