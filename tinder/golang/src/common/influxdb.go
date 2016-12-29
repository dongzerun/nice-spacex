package common

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/influxdata/influxdb/client"
)

// 通用的 influxdb 封装
// 1. 首先调用 InitHandler 初始化全局的 handler, 类似日志一样
// 2. 调用 PutMetric 或 PutMetricLazy 开始灌数据。区别在于，
//  Lazy 是惰性，等待聚合写入，一般建义用这个，减少 influxdb 压力
//
//            hostname, port,  database,   user,    password,   buffer size
// InitHandler(h string, p int, db string, u string, pwd string, max int)
//
// 建义使用 PutMetricLazy
// PutMetricLazy(table string, tags map[string]string, data map[string]interface{})
// table 就是统计要入的 Measurement Name
// Tags 是一个map，可以理解为 MySQL 中建了特殊索引的列，这些列是可以做 GROUP BY的
// Tags 的值必须是 string, 且不能出现在 Data 的map里
// Tags: map[string]string{
// 	"hostname": Hostname,
// 	"action":   k,
// },
// Data 是一个map，key 是 string, 但是value可以是任何类型
// Data 中的内容不能出现在 Tags 中
// Fields: map[string]interface{}{
// 	"count": value,
// },

var (
	InfluxdbBufferFull = errors.New("influxdb buffer full")
	HandlerUnInited    = errors.New("influxdb handler not inited ")
)

var ih *InfluxHandler

type unit struct {
	table string
	tags  map[string]string
	data  map[string]interface{}
	t     time.Time
}

type InfluxHandler struct {
	c   *client.Client
	h   string
	p   int
	db  string
	u   string
	pwd string

	max   int
	ch    chan unit
	force chan struct{}

	last time.Time
}

func NewInfluxHandler(h string, p int, db string, u string, pwd string, max int) (*InfluxHandler, error) {
	url, err := client.ParseConnectionString(fmt.Sprintf("%s:%d", h, p), false)
	if err != nil {
		return nil, err
	}

	cfg := client.NewConfig()
	cfg.URL = url
	cfg.Username = u
	cfg.Password = pwd
	cfg.Precision = "s" // 秒级别的精度就行

	c, e := client.NewClient(cfg)
	if e != nil {
		return nil, e
	}

	_, _, err = c.Ping()
	if err != nil {
		return nil, err
	}

	ih := new(InfluxHandler)
	ih.c = c
	ih.force = make(chan struct{}, 1)
	ih.last = time.Now()
	ih.h = h
	ih.p = p
	ih.db = db
	ih.u = u
	ih.pwd = pwd

	if max <= 0 || max > 4096 {
		ih.max = 4096
	} else {
		ih.max = max
	}

	ih.ch = make(chan unit, int(ih.max+ih.max/2))

	go ih.consume()

	return ih, nil
}

// max 是最大聚合数量
func InitHandler(h string, p int, db string, u string, pwd string, max int) error {
	var err error
	ih, err = NewInfluxHandler(h, p, db, u, pwd, max)
	return err
}

func PutMetric(table string, tags map[string]string, data map[string]interface{}) error {
	if ih == nil {
		return HandlerUnInited
	}
	return ih.PutMetric(table, tags, data)
}

func PutMetricLazy(table string, tags map[string]string, data map[string]interface{}) error {
	if ih == nil {
		return HandlerUnInited
	}
	return ih.PutMetricLazy(table, tags, data)
}

// tags 里面有的字段，不能在data里
func (h *InfluxHandler) PutMetric(table string, tags map[string]string, data map[string]interface{}) error {
	err := h.putMetric(table, tags, data)

	// 写入metric后，立马通知刷新
	select {
	case h.force <- struct{}{}:
	default:
	}
	return err
}

func (h *InfluxHandler) PutMetricLazy(table string, tags map[string]string, data map[string]interface{}) error {
	return h.putMetric(table, tags, data)
}

func (h *InfluxHandler) putMetric(table string, tags map[string]string, data map[string]interface{}) error {
	u := unit{
		table: table,
		tags:  tags,
		data:  data,
		t:     time.Now(),
	}

	select {
	case h.ch <- u:
	default:
		return InfluxdbBufferFull
	}
	return nil
}

func (h *InfluxHandler) consume() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Println("InfluxHandler consume panic ", string(buf))
		}
	}()

	var (
		pts   = make([]client.Point, 0, h.max+1)
		flush bool
		idx   = 0
	)

	for {
		select {
		case u := <-h.ch:
			p := client.Point{
				Measurement: u.table,
				Time:        u.t, // 时间必须为send到ch的时间
				Precision:   "s",
				Tags:        u.tags,
				Fields:      u.data,
			}

			pts = append(pts, p)

			idx++
			// 索引回卷
			// idx = idx % h.max
		case <-h.force:
			// 收到强制刷新信号
			flush = true
			// 创建太多的定时器性能低，influxdb本身就不用太精确移到外面if语句就好
			// case <-time.After(600 * time.Second):
			// 	if idx > 0 {
			// 		flush = true
			// 	}
		}

		// 有未刷数据，并且距离上一次刷数据超过 10min
		if time.Now().Sub(h.last) > 600*time.Second && idx > 0 {
			flush = true
		}

		// 未刷数据量，超过了自定义上限
		if idx >= h.max {
			flush = true
		}

		// 刷新合并后的数据 >> influxdb
		if flush {
			err := h.writeToInfluxdb(pts[:idx])
			if err != nil {
				fmt.Println("write to influxdb err ", err.Error())
			}
			pts = pts[:0]
			idx = 0
			flush = false
			h.last = time.Now()
		}
	}
}

func (h *InfluxHandler) writeToInfluxdb(pts []client.Point) error {
	bps := client.BatchPoints{
		Points:          pts,
		Database:        h.db,
		RetentionPolicy: "default",
	}

	_, err := h.c.Write(bps)
	return err
}
