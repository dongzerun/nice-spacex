package wrapad

import (
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	tf "spacex_tf/spacex"
	"xad"
	"xuser"
	"xutil/cache"
)

func init() {
	RegisterPostFunc("time_jitter", TimeJitter)
}

// 为了防止出现抖动现象，对于同一个人同一广告，连续5s内请求只计数一次
func TimeJitter(ctx context.Context, ad *xad.Ad, req *tf.RequestParams, source string) {
	if source != "GetAd" {
		return
	}
	key := fmt.Sprintf(defined.KeyTimeJitterFormat, ad.Id, *req.UID)
	exists, err := cache.GlobalRedisCache.Exists(key)
	if err != nil {
		common.Warningf("logid:%v TimeJitter check %s exists error", ctx.Value("logid"), key)
		return
	}

	if exists {
		//回滚 display num 计数
		ui := &xuser.UserInfo{
			Uid: *req.UID,
		}

		common.Warningf("logid:%v TimeJitter start rollback %s %v", ctx.Value("logid"), key, exists)
		ad.RollBackDisplayNum(ctx, ui)
		return
	}

	// 不存在的话，设置key，并设置超时时间2s
	common.Warningf("logid:%v TimeJitter set %s ttl 2s", ctx.Value("logid"), key)
	cache.GlobalRedisCache.SetByteWithEx(key, []byte("jitter"), 2)
	return
}
