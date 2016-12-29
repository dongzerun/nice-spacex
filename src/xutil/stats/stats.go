package command

import (
	"os"
	"strconv"
	"strings"
	"time"

	"common"
	"defined"

	tf "spacex_tf/spacex"
)

var (
	// 响应类型channel
	statsCh chan *tf.Response

	// 本机 hostname
	Hostname string

	// 广告业务统计
	s *statsSpacex

	// 响应时间分类统计
	elapse *timeConsume
	// 响应时间 channel
	duration chan int64
)

func init() {
	statsCh = make(chan *tf.Response, 4096)
	s = &statsSpacex{
		response: make(map[string]int64),
		ads:      make(map[int]adStats),
	}

	elapse = &timeConsume{}
	duration = make(chan int64, 4096)
	go statsLoop()

	host, err := os.Hostname()
	if err != nil {
		panic("get hostname failed " + err.Error())
	}
	Hostname = strings.Split(host, ".")[0]
}

type statsSpacex struct {
	qps      int64
	response map[string]int64
	ads      map[int]adStats
}

type adStats struct {
	id     int
	name   string
	area   string
	adtype string
	v      int64
}

// 小于 1ms
// 1ms ~ 10ms
// 10ms ~ 50ms
// 50ms ~ 100ms
// 100ms ~200ms
// 200ms ~500ms
// 500ms ~1s
// 1s ~ 2s
type timeConsume struct {
	ms1      int64
	ms10     int64
	ms50     int64
	ms100    int64
	ms150    int64
	ms200    int64
	ms500    int64
	ms1000   int64
	ms2000   int64
	msgt2000 int64
}

// 接受响应时间参数，单位是 us 微妙
func (tc *timeConsume) Putelapse(us int64) {
	switch {
	case us <= 1000: // 小于 1ms
		tc.ms1 += 1
	case 1000 < us && us <= 10000: // 1ms ~ 10ms
		tc.ms10 += 1
	case 10000 < us && us <= 50000: // 10ms ~ 50ms
		tc.ms50 += 1
	case 50000 < us && us <= 100000: // 50ms ~ 100ms
		tc.ms100 += 1
	case 100000 < us && us <= 150000: // 100ms ~ 150ms
		tc.ms150 += 1
	case 150000 < us && us <= 200000: // 150ms ~ 200ms
		tc.ms200 += 1
	case 200000 < us && us <= 500000: // 200ms ~ 500ms
		tc.ms500 += 1
	case 500000 < us && us <= 1000000: // 500ms ~ 1s
		tc.ms1000 += 1
	case 1000000 < us && us <= 2000000: // 1s ~ 2s
		tc.ms2000 += 1
	case 2000000 < us: // > 2s
		tc.msgt2000 += 1
	}
}

func PutResponseStats(res *tf.Response) {
	select {
	case statsCh <- res:
	default:
		common.Warning("spacex cmd stats channel full")
	}
}

func PutDuration(d int64) {
	select {
	case duration <- d:
	default:
		common.Warning("spacex cmd duration channel full")
	}
}

func PushDuration(elapse *timeConsume) {
	pushDuration("min~1ms", elapse.ms1)
	pushDuration("1ms~10ms", elapse.ms10)
	pushDuration("10ms~50ms", elapse.ms50)
	pushDuration("50ms~100ms", elapse.ms100)
	pushDuration("100ms~150ms", elapse.ms150)
	pushDuration("150ms~200ms", elapse.ms200)
	pushDuration("200ms~500ms", elapse.ms500)
	pushDuration("500ms~1s", elapse.ms1000)
	pushDuration("1s~2s", elapse.ms2000)
	pushDuration("2s~max", elapse.msgt2000)
}

func pushDuration(name string, num int64) {
	tags := map[string]string{
		"hostname": Hostname,
		"duration": name,
	}
	data := map[string]interface{}{
		"count": num,
	}

	err := common.PutMetric("elapse", tags, data)
	if err != nil {
		common.Warning("statsLoop PutMetric err ", err.Error())
	}
}

func PushStats(as *statsSpacex) {
	// 将 QPS 打入 influxdb qps 表
	tags := map[string]string{
		"hostname": Hostname,
	}
	data := map[string]interface{}{
		"count": as.qps,
	}

	err := common.PutMetric("qps", tags, data)
	if err != nil {
		common.Warning("statsLoop PutMetric err ", err.Error())
	}

	for status, _ := range as.response {
		// 将返回状态打入 influxdb repsonse 表
		tags := map[string]string{
			"status": status,
		}
		data := map[string]interface{}{
			"count": as.response[status],
		}

		err := common.PutMetric("response", tags, data)
		if err != nil {
			common.Warning("statsLoop PutMetric err ", err.Error())
		}
	}

	for _, ad := range as.ads {
		tags := map[string]string{
			"adid":   strconv.Itoa(int(ad.id)),
			"name":   ad.name,
			"area":   ad.area,
			"adtype": ad.adtype,
		}
		data := map[string]interface{}{
			"count": ad.v,
		}

		err := common.PutMetric("adstats", tags, data)
		if err != nil {
			common.Warning("statsLoop PutMetric err ", err.Error())
		}
	}
}

func statsLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case d := <-duration:
			elapse.Putelapse(d)
		case res := <-statsCh:
			s.qps += 1
			s.response[*res.ErrorString] += 1

			// 状态正常，将广告信息打入到 influxdb 统计中
			if res.Status != defined.StatusOK || res.Data == nil {
				continue
			}

			for _, ad := range res.Data.Ads {
				as, ok := s.ads[int(ad.AdId)]
				if !ok {
					s.ads[int(ad.AdId)] = adStats{
						id:     int(ad.AdId),
						name:   ad.AdName,
						area:   ad.AdArea,
						adtype: ad.AdType,
						v:      1,
					}
				} else {
					as.v += 1
				}
			}
		case <-ticker.C:
			old := s
			s = &statsSpacex{
				response: make(map[string]int64),
				ads:      make(map[int]adStats),
			}

			go PushStats(old)

			e := elapse
			elapse = &timeConsume{}

			go PushDuration(e)
		}
	}
	common.Warning("Spacex stats quit statsLoop goroutine")
}
