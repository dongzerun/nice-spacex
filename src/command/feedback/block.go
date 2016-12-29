package feedback

import (
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	tf "spacex_tf/spacex"
	"xad"
	"xutil/cache"
	"xutil/hack"
	stats "xutil/stats"
)

var _ Action = (*Block)(nil)

// 注册 block 屏蔽广告处理函数
func init() {
	Register("block", func() Action {
		b := new(Block)
		b.SetName("block")
		return b
	})
}

type Block struct {
	Base
}

// 屏蔽广告信息，那就是在Redis里设置超时key，ttl为广告剩余有效时间
// set key ttl elapse time
func (bl *Block) Process(ctx context.Context, req *tf.FeedbackParams) *tf.Response {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	ttl, err := xad.AdM.Remaining(ctx, int(req.AdId))
	if err != nil {
		common.Warningf("logid:%v get adid: %d Remaining err:%s", ctx.Value("logid"), req.AdId, err.Error())
	}

	if ttl <= 0 {
		ttl = defined.TTLBlockActionDefault
	}

	key := fmt.Sprintf(defined.KeyBlockActionFormat, req.UID, req.AdId)

	err = cache.GlobalRedisCache.SetByteWithEx(key, hack.Slice(""), int64(ttl))

	if err != nil {
		common.Warningf("logid:%v SetByteWithEx err:%s", ctx.Value("logid"), err.Error())
		res := defined.GetResponse(defined.StatusUnknownError)
		stats.PutResponseStats(res)
		return res
	}

	common.Warningf("logid:%v SetByteWithEx ttl:%d", ctx.Value("logid"), ttl)

	res := defined.GetResponse(defined.StatusOK)
	stats.PutResponseStats(res)
	return res
}
