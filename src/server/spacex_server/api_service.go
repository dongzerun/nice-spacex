package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/ant0ine/go-json-rest/rest"

	"common"
	"defined"
	"xad/xplugin"
)

type AdFeedBack struct {
	Action string `json:"action"`
	Uid    int64  `json:"uid"`
	AdId   int    `json:"ad_id"`
}

func StartApiServer(port int) {
	go func() {
		api := rest.NewApi()
		api.Use(rest.DefaultDevStack...)
		router, err := rest.MakeRouter(
			rest.Get("/reload", Reload),
			rest.Get("/system", GetSystemMetrics),
			rest.Post("/feedback/display", FeedbackDisplay),
		)
		if err != nil {
			common.Error(err)
		}
		api.SetApp(router)
		common.Info("start api server @", port)
		err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), api.MakeHandler())
		if err != nil {
			common.Error(err)
		}
	}()
}

// 获取metric参数
func GetSystemMetrics(w rest.ResponseWriter, r *rest.Request) {
	routineNum := runtime.NumGoroutine()
	memStats := &runtime.MemStats{}

	runtime.ReadMemStats(memStats)

	ret := make(map[string]interface{})

	ret["routine.nums"] = routineNum
	ret["mem.Alloc"] = memStats.Alloc
	ret["mem.TotalAlloc"] = memStats.TotalAlloc
	ret["mem.Sys"] = memStats.Sys
	ret["mem.Lookups"] = memStats.Lookups
	ret["mem.Mallocs"] = memStats.Mallocs
	ret["mem.Frees"] = memStats.Frees

	ret["mem.heap.HeapAlloc"] = memStats.HeapAlloc
	ret["mem.heap.HeapSys"] = memStats.HeapSys
	ret["mem.heap.HeapIdle"] = memStats.HeapIdle
	ret["mem.heap.HeapInuse"] = memStats.HeapInuse
	ret["mem.heap.HeapReleased"] = memStats.HeapReleased
	ret["mem.heap.HeapObjects"] = memStats.HeapObjects

	ret["mem.stack.StackInuse"] = memStats.StackInuse
	ret["mem.stack.StackSys"] = memStats.StackSys
	ret["mem.stack.MSpanInuse"] = memStats.MSpanInuse
	ret["mem.stack.MSpanSys"] = memStats.MSpanSys
	ret["mem.stack.MCacheInuse"] = memStats.MCacheInuse
	ret["mem.stack.MCacheSys"] = memStats.MCacheSys
	ret["mem.stack.BuckHashSys"] = memStats.BuckHashSys
	ret["mem.stack.GCSys"] = memStats.GCSys
	ret["mem.stack.OtherSys"] = memStats.OtherSys

	ret["mem.gc.NextGC"] = memStats.NextGC
	ret["mem.gc.LastGC"] = memStats.LastGC
	ret["mem.gc.PauseTotalNs"] = memStats.PauseTotalNs
	ret["mem.gc.PauseNs"] = memStats.PauseNs
	ret["mem.gc.NumGC"] = memStats.NumGC
	ret["mem.gc.EnableGC"] = memStats.EnableGC
	ret["mem.gc.DebugGC"] = memStats.DebugGC
	w.WriteJson(ret)
}

// 设置阈值
func Reload(w rest.ResponseWriter, r *rest.Request) {
	ret := make(map[string]interface{})
	ret["status"] = 200

	defer func() {
		if err := recover(); err != nil {
			ret["status"] = 400
			w.WriteJson(ret)
		}
	}()

	defined.AddEvent(defined.E_RELOAD_AD)

	common.Info("reload all configuration")
	w.WriteJson(ret)
}

func FeedbackDisplay(w rest.ResponseWriter, r *rest.Request) {
	defer func() {
		if e := recover(); e != nil {
			rest.Error(w, "SERVER INTERNAL ERROR", http.StatusInternalServerError)
		}
	}()

	if r.Method != "POST" {
		rest.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fb := new(AdFeedBack)

	err := r.DecodeJsonPayload(fb)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	common.Infof("FeedbackDisplay uid:%d, ad:%d action:%s", fb.Uid, fb.AdId, fb.Action)

	if fb.Action != "display" {
		common.Warningf("FeedbackDisplay wanted action display but get %s", fb.Action)
		rest.Error(w, "feedback action not display", http.StatusBadRequest)
		return
	}

	xplugin.IncrCntByAdV2(fb.AdId, fb.Uid)

	ret := make(map[string]interface{})
	ret["status"] = 200
	w.WriteJson(ret)
}
