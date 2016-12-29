package command

import (
	"encoding/json"
	"runtime"

	"golang.org/x/net/context"

	service "command/service"
	"common"
	"defined"
	tf "spacex_tf/spacex"
	"time"
	"xad"   // 广告模块
	"xuser" // 用户模块
	"xutil/hack"
	stats "xutil/stats"
)

func Dispatch(logId string, req *tf.RequestParams) *tf.Response {

	bytes, _ := json.Marshal(req)

	common.Infof("logid:%s request:%s", logId, hack.String(bytes))

	var uid int64

	uid = *req.UID
	// uid 必须存在, 要从redis cache中找到 UidLastAdKeyFormat
	if uid == 0 && req.Method != "enter_screen" && req.Method != "live_h5" {
		res := defined.GetResponse(defined.StatusReqUidNull)
		stats.PutResponseStats(res)
		return res
	}

	fn := func(ctx context.Context, myAd chan *tf.Response) {
		defer func() {
			// 捕获 当前goroutine的error
			if err := recover(); err != nil {
				buf := make([]byte, defined.STACKSIZE)
				buf = buf[:runtime.Stack(buf, false)]
				common.Errorf("logid:%v SpaceX Dispatch panic %s", ctx.Value("logid"), string(buf))
			}
		}()
		var (
			cmd service.Command
			ok  bool
			ads []*xad.Ad
			ui  *xuser.UserInfo
			err error
		)

		if cmd, ok = service.CommandSerivce[req.Method]; !ok {
			common.Errorf("logid:%v Not Found request method", ctx.Value("logid"))
			res := defined.GetResponse(defined.StatusNotFoundMethod)
			stats.PutResponseStats(res)
			myAd <- res
			return
		}

		// ui = new(xuser.UserInfo) 无用代码
		if req.UID != nil {
			// 获取用户内容
			// ui, err = new(xuser.User).GetUserInfo(ctx, uid)
			// u := new(xuser.User)
			u := xuser.PickUserGetter()
			defer xuser.PutUserGetter(u)

			common.Debugf("logid:%v before run get user info has spent:%v", ctx.Value("logid"), time.Since(ctx.Value("start_time").(time.Time)).Nanoseconds()/1000)
			ui, err = u.GetUserInfo(ctx, uid)
			common.Debugf("logid:%v after run get user info has spent:%v", ctx.Value("logid"), time.Since(ctx.Value("start_time").(time.Time)).Nanoseconds()/1000)
			if err != nil {
				common.Errorf("logid:%v Get %d user info error %s", ctx.Value("logid"), uid, err.Error())
				res := defined.GetResponse(defined.StatusMissUserInfo)
				stats.PutResponseStats(res)
				myAd <- res
				return
			}
			defer u.PutUserInfo(ui)
		}

		ui.Req = req

		ads, err = cmd.Match(ctx, ui)
		if err != nil || len(ads) == 0 {
			common.Infof("logid:%v uid:%d did:%s match failure %s", ctx.Value("logid"), uid, req.DeviceID, err.Error())
			res := defined.GetResponse(defined.StatusMissMatchAd)
			stats.PutResponseStats(res)
			myAd <- res
			return
		}

		saveid := ads[len(ads)-1].Id

		err = cmd.SaveAdUid(ctx, *req.UID, saveid)
		if err != nil {
			common.Warningf("logid:%v cmd save ad uid failed %s", ctx.Value("logid"), err.Error())
		}

		res := WrapRes(ctx, req, "GetAd", ads...)

		stats.PutResponseStats(res)

		myAd <- res
	}

	return ExecCommand(logId, req.Timeout, fn)
}

func GetAdByUid(logId string, req *tf.RequestParams) *tf.Response {

	bytes, _ := json.Marshal(req)

	common.Infof("logid:%s request:%s", logId, hack.String(bytes))

	var uid int64

	uid = *req.UID
	// uid 必须存在, 要从redis cache中找到 UidLastAdKeyFormat
	if uid == 0 {
		res := defined.GetResponse(defined.StatusReqUidNull)
		stats.PutResponseStats(res)
		return res
	}

	fn := func(ctx context.Context, myAd chan *tf.Response) {
		defer func() {
			// 捕获 当前goroutine的error
			if err := recover(); err != nil {
				buf := make([]byte, defined.STACKSIZE)
				buf = buf[:runtime.Stack(buf, false)]
				common.Errorf("logid:%v SpaceX GetAdByUid panic ", ctx.Value("logid"), string(buf))
			}
		}()

		var (
			cmd service.Command
			ok  bool
			err error
			ad  *xad.Ad
		)

		if cmd, ok = service.CommandSerivce[req.Method]; !ok {
			common.Errorf("logid:%v Not Found request method %s", ctx.Value("logid"), req.Method)
			res := defined.GetResponse(defined.StatusNotFoundMethod)
			stats.PutResponseStats(res)
			myAd <- res
			return
		}

		ad, err = cmd.GetAdByUid(ctx, uid)
		if err != nil {
			common.Warningf("logid:%v GetAdByUid uid:%d err:%s", ctx.Value("logid"), uid, err.Error())
			res := defined.GetResponse(defined.StatusAdIdNonExists)
			stats.PutResponseStats(res)
			myAd <- res
			return
		}

		res := WrapRes(ctx, req, "GetAdByUid", ad)

		stats.PutResponseStats(res)
		myAd <- res
	}

	return ExecCommand(logId, req.Timeout, fn)
}

func GetAdByAdId(logId string, req *tf.GetAdParams) *tf.Response {

	bytes, _ := json.Marshal(req)

	common.Infof("logid:%s request:%s", logId, hack.String(bytes))

	fn := func(ctx context.Context, myAd chan *tf.Response) {
		defer func() {
			// 捕获 当前goroutine的error
			if err := recover(); err != nil {
				buf := make([]byte, defined.STACKSIZE)
				buf = buf[:runtime.Stack(buf, false)]
				common.Errorf("logid:%v SpaceX GetAdByAdId panic ", ctx.Value("logid"), string(buf))
			}
		}()

		var (
			err error
			ad  *xad.Ad
		)

		ad, err = xad.AdM.GetAd(int(req.AdId))
		if err != nil {
			common.Warningf("logid:%v GetAdByAdId uid:%d err:%s", ctx.Value("logid"), req.UID, err.Error())
			res := defined.GetResponse(defined.StatusAdIdNonExists)
			stats.PutResponseStats(res)
			myAd <- res
			return
		}

		res := WrapResDirect(ctx, req.UID, "GetAdByAdId", ad)

		stats.PutResponseStats(res)
		myAd <- res
	}

	return ExecCommand(logId, req.Timeout, fn)
}
