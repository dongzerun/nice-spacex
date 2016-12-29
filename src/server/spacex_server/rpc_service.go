package main

import (
	"encoding/json"
	"fmt"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/sluu99/uuid"

	"command"
	"common"
	tf "spacex_tf/spacex"
	"xutil/hack"
	stats "xutil/stats"
)

type TFServer struct {
}

// 线上各个广告位匹配广告核心 API
func (this *TFServer) GetAd(req *tf.RequestParams) (res *tf.Response, err error) {

	startTime := time.Now()
	logId := LogID(req.Logid)
	res = command.Dispatch(logId, req)
	cost := time.Since(startTime).Nanoseconds() / 1000
	bytes, _ := json.Marshal(res)
	stats.PutDuration(cost)
	common.Infof("logid:%s GetAd Method:%s response result:%s cost:%dus", logId, req.Method, hack.String(bytes), cost)

	return res, nil
}

// 获取用户最后一次广告详情
func (this *TFServer) GetAdDetail(req *tf.RequestParams) (res *tf.Response, err error) {

	startTime := time.Now()
	logId := LogID(req.Logid)
	res = command.GetAdByUid(logId, req)
	cost := time.Since(startTime).Nanoseconds() / 1000
	bytes, _ := json.Marshal(res)
	stats.PutDuration(cost)
	common.Infof("logid:%s GetAdDetail Method:%s response result:%s cost:%dus", logId, req.Method, hack.String(bytes), cost)

	return res, nil
}

// 点击反馈
func (this *TFServer) FeedBack(req *tf.FeedbackParams) (res *tf.Response, err error) {
	startTime := time.Now()
	logId := LogID(req.Logid)
	res = command.FeedBack(logId, req)
	cost := time.Since(startTime).Nanoseconds() / 1000
	bytes, _ := json.Marshal(res)
	stats.PutDuration(cost)
	common.Infof("logid:%s FeedBack Action:%s response result:%s cost:%dus", logId, req.Action, hack.String(bytes), cost)

	return res, nil
}

// 根据adid获取广告详情 API
func (this *TFServer) GetAdByAdId(req *tf.GetAdParams) (res *tf.Response, err error) {
	startTime := time.Now()
	logId := LogID(req.Logid)
	res = command.GetAdByAdId(logId, req)
	cost := time.Since(startTime).Nanoseconds() / 1000
	bytes, _ := json.Marshal(res)
	stats.PutDuration(cost)
	common.Infof("logid:%s GetAdByAdId response result:%s cost:%dus", logId, hack.String(bytes), cost)

	return res, nil
}

// 返回 log id,传进为空，那么默认返回 uuid
func LogID(id string) string {
	if id != "" {
		return id
	}
	return uuid.Rand().Hex()
}

func StartAdServer(port int) {
	//transportFactory := thrift.NewTTransportFactory()
	transportFactory := thrift.NewTBufferedTransportFactory(10240)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	serverTransport, err := thrift.NewTServerSocket(fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	common.Infof("thrift rpc listen %s ", fmt.Sprintf(":%d", port))

	handler := new(TFServer)
	processor := tf.NewGreeterProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	server.Serve()

}
