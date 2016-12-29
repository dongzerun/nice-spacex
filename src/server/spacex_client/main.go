package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
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
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	runtime.GOMAXPROCS(runtime.NumCPU())

	//tSocket, err := thrift.NewTSocketTimeout(fmt.Sprintf("%s:%d", "10.10.10.198", 50049), time.Second)
	tSocket, err := thrift.NewTSocket(fmt.Sprintf("%s:%d", "localhost", 7680))

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
		uid int64 = 18555476
		//uid      int64  = 0
		timeout  int64  = 1000
		deviceId string = "d65db594510aef4468694c14d0970f1e"
		logid    string = "2342342342"
		//method   string = "enter_screen"
		//method string = "feed_card_3"
		//method string = "tag_link"
		//method string = "editor_recommend"
		//method string = "paster_detail"
		method string = "vsfeed_card_3"

		channel     string = "AppStore_3.8.2.1"
		device_type string = "iPhone7.2"
		device_os   string = "9.2"
		net         string = "0-0-wifi"
		app_version string = "3.10.10"

		lat  string = "39.953810"
		lang string = "116.453964"

		tag_id   string = "23"
		tag_type string = "exists"

		package_id string = "23"
	)

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
			TagID:      &tag_id,
			TagType:    &tag_type,
			PackageID:  &package_id,
		},
	}

	res, err := client.GetAd(req)
	fmt.Println(res, err)
	//fmt.Printf("%#v, %#v\n", res, err)
}
