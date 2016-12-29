package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	tf "spacex_tf/spacex"
)

var (
	parallel = flag.Int("parallel", 1, "benchmark parallel")
	totalnum = flag.Int64("totalnum", 1, "benchmark total number")
	loop     = flag.Bool("loop", false, "benchmark loop forever, override totalsum")
	host     = flag.String("host", "", "service hostname")
	port     = flag.Int("port", 0, "service port")
	users    = flag.String("userfile", "", "active user list, used to benchmark")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	runtime.GOMAXPROCS(runtime.NumCPU())

	if *parallel <= 0 || *totalnum <= 0 {
		flag.PrintDefaults()
		log.Fatal("input arg Illegal")
	}

	if *host == "" || *port == 0 {
		log.Fatal("host or port must not empty")
	}

	var qps int64

	areas := []string{"enter_screen", "feed_card_3", "tag_link"}

	last := NewTSQ(100000)

	quit := make(chan struct{}, 1)

	activeUser := make([]int64, 0, 20567462)

	if users != nil {
		data, err := ioutil.ReadFile(*users)
		if err == nil {
			slices := strings.Split(string(data), "\n")
			if len(slices) != 0 {
				for _, v := range slices {
					id, e := strconv.ParseInt(v, 10, 64)
					if e == nil {
						activeUser = append(activeUser, id)
					}
				}
			}
		}
	}

	if len(activeUser) == 0 {
		for i := 0; i <= 20567462; i++ {
			activeUser = append(activeUser, int64(i))
		}
	}

	for i := 0; i < *parallel; i++ {
		go func() {

			//tSocket, err := thrift.NewTSocketTimeout(fmt.Sprintf("%s:%d", "10.10.10.198", 50049), time.Second)
			tSocket, err := thrift.NewTSocketTimeout(fmt.Sprintf("%s:%d", *host, *port), 40*time.Millisecond)

			if err != nil {
				log.Fatal("thrift NewTSocket failed ", err)
			}

			transportFactory := thrift.NewTTransportFactory()
			transport := transportFactory.GetTransport(tSocket)

			protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
			client := tf.NewGreeterClientProtocol(transport, protocolFactory.GetProtocol(transport), protocolFactory.GetProtocol(transport))

			if err = transport.Open(); err != nil {
				log.Print("tSocket open failed ", err)
				return
			}

			var (
				uid      int64
				timeout  int64  = 1000
				deviceId string = "d65db594510aef4468694c14d0970f1e"
				logid    string = "2342342342"
				method   string

				channel     string = "AppStore_3.8.2.1"
				device_type string = "iPhone7.2"
				device_os   string = "9.2"
				net         string = "0-0-wifi"
				app_version string = "3.10.10"

				lat  string = "39.953810"
				lang string = "116.453964"
			)

			for {
				uid = activeUser[rand.Int31n(int32(len(activeUser)))]

				method = areas[rand.Int31n(3)]

				req := &tf.RequestParams{
					UID:      &uid,
					DeviceID: deviceId,
					Logid:    logid,
					Method:   method,
					Timeout:  &timeout,
					Extra: &tf.Extra{
						Channel:    &channel,
						DeviceType: &device_type,
						DeviceOs:   &device_os,
						Net:        &net,
						AppVersion: &app_version,
						Latitude:   &lat,
						Longitude:  &lang,
					},
				}

				now := time.Now().UnixNano()
				_, err := client.GetAd(req)
				if err != nil {
					fmt.Println("GetAd err ", err.Error())
				}
				// client.GetAd(req)
				// fmt.Println(res, err)
				// var e string
				// if res.ErrorString != nil {
				// 	e = *res.ErrorString
				// }

				// if res.Data != nil {
				// 	fmt.Println(res.Status, e, res.Data, err)
				// }

				atomic.AddInt64(&qps, 1)

				consumed := time.Now().UnixNano() - now

				if rand.Intn(10) == 1 {
					last = append(last, consumed)
					sort.Sort(last)
				}

				if !*loop && qps > *totalnum {
					quit <- struct{}{}
					return
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Printf("QPS: %d", qps/10)
				atomic.StoreInt64(&qps, 0)
				old := last
				last = NewTSQ(100000)

				var avg99 int64 = 0
				var avg95 int64 = 0
				var total int64 = 0
				for i, _ := range old {
					total += old[i]
					if i == (len(old) * 99 / 100) {
						avg99 = total / int64(i+1)
					}
					if i == (len(old) * 95 / 100) {
						avg95 = total / int64(i+1)
					}
				}
				if len(old) == 0 {
					continue
				}

				log.Printf("avg: %dms, avg99: %dms, avg95: %dms, max: %dms, min: %dms", total/int64(len(old))/1000/1000, avg99/1000/1000, avg95/1000/1000, old[len(old)-1]/1000/1000, old[0]/1000/1000)
			}
		}
	}()

	<-quit
}

type TimeStatsQueue []int64

func NewTSQ(capacity int) TimeStatsQueue {
	return make(TimeStatsQueue, 0, capacity)
}

func (ts TimeStatsQueue) Len() int {
	return len(ts)
}

func (ts TimeStatsQueue) Less(i, j int) bool {
	return ts[i] < ts[j]
}

func (ts TimeStatsQueue) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}
