package command

import (
	"encoding/json"
	"time"

	"golang.org/x/net/context"

	"command/wrapad"
	"common"
	"config"
	"defined"
	tf "spacex_tf/spacex"
	"xad"
	"xutil/hack"
	stats "xutil/stats"
)

// t 超时时间单位是ms的指针，fn传入参数为*tf.Response的channel
func ExecCommand(logid string, t *int64, fn func(context.Context, chan *tf.Response)) *tf.Response {
	timeout := time.Millisecond * time.Duration(config.GlobalConfig.SConfig.DefaultTimeout)
	if t != nil {
		// 取客户端超时的 8/10 做为服务的超时时间
		timeout = time.Millisecond * time.Duration((*t)*8/10)
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	ctx = context.WithValue(ctx, "start_time", time.Now())

	r := make(chan *tf.Response, 1)

	go fn(context.WithValue(ctx, "logid", logid), r)

	select {
	case <-ctx.Done():
		res := defined.GetResponse(defined.StatusTimeout)
		stats.PutResponseStats(res)
		return res
	case rr := <-r:
		return rr
	}

	res := defined.GetResponse(defined.StatusUnknownError)
	stats.PutResponseStats(res)
	return res
}

func WrapRes(ctx context.Context, req *tf.RequestParams, source string, ads ...*xad.Ad) *tf.Response {

	res := defined.GetResponse(defined.StatusOK)
	res.Data = &tf.DataBag{
		UID: *req.UID,
		Ads: make([]*tf.Ad, 0, 1),
	}

	for _, ad := range ads {
		// 对ad做一次封装操作
		ad = wrapAd(ctx, ad, req, source)

		bytes, err := json.Marshal(ad)
		if err != nil {
			common.Errorf("logid:%v encode ad content error ", ctx.Value("logid"), ad)

			r := defined.GetResponse(defined.StatusUnknownError)
			stats.PutResponseStats(r)
			return r
		}

		tad := &tf.Ad{
			AdId:     int32(ad.Id),
			AdName:   ad.Name,
			AdArea:   ad.Area,
			AdType:   ad.Type,
			AdDetail: hack.String(bytes),
		}
		res.Data.Ads = append(res.Data.Ads, tad)
	}
	return res
}

func WrapResDirect(ctx context.Context, uid int64, source string, ads ...*xad.Ad) *tf.Response {

	res := defined.GetResponse(defined.StatusOK)
	res.Data = &tf.DataBag{
		UID: uid,
		Ads: make([]*tf.Ad, 0, 1),
	}

	for _, ad := range ads {

		bytes, err := json.Marshal(ad)
		if err != nil {
			common.Errorf("logid:%v encode ad content error ", ctx.Value("logid"), ad)

			r := defined.GetResponse(defined.StatusUnknownError)
			stats.PutResponseStats(r)
			return r
		}

		tad := &tf.Ad{
			AdId:     int32(ad.Id),
			AdName:   ad.Name,
			AdArea:   ad.Area,
			AdType:   ad.Type,
			AdDetail: hack.String(bytes),
		}
		res.Data.Ads = append(res.Data.Ads, tad)
	}
	return res
}

// 对返回的广告做一次封装，可能会返回原始广告，或是返回修改后的
func wrapAd(ctx context.Context, ad *xad.Ad, req *tf.RequestParams, source string) *xad.Ad {

	//做深度拷贝, 不修改原始 ad 数据
	newAd := xad.DeepCopyAd(ad)

	// 轮流执行注册的回调函数
	for name, callback := range wrapad.PostFuncs {
		common.Infof("wrapAd call %s function", name)
		callback(ctx, newAd, req, source)
	}

	//TODO:
	//暂时只返回老的ad，等调试通过后边再返回新的
	return newAd
}
