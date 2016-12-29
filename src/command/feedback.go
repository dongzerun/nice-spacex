package command

import (
	"encoding/json"
	"runtime"

	"golang.org/x/net/context"

	feedback "command/feedback"
	"common"
	"defined"
	tf "spacex_tf/spacex"
	"xutil/hack"
	stats "xutil/stats"
)

func FeedBack(logId string, req *tf.FeedbackParams) *tf.Response {
	bytes, _ := json.Marshal(req)

	common.Infof("logid:%s request:%s", logId, hack.String(bytes))

	uid := req.UID
	// uid 必须存在, 要从redis cache中找到 UidLastAdKeyFormat
	if uid == 0 {
		res := defined.GetResponse(defined.StatusReqUidNull)
		stats.PutResponseStats(res)
		return res
	}

	fn := func(ctx context.Context, r chan *tf.Response) {
		defer func() { // 捕获 当前goroutine的error
			if err := recover(); err != nil {
				buf := make([]byte, defined.STACKSIZE)
				buf = buf[:runtime.Stack(buf, false)]
				common.Errorf("logid:%v SpaceX FeedBack panic ", ctx.Value("logid"), string(buf))
			}
		}()

		a, ok := feedback.FeedServiceMap[req.Action]
		if !ok {
			res := defined.GetResponse(defined.StatusActionEmptyOrIllegal)
			stats.PutResponseStats(res)
			r <- res
			return
		}

		r <- a.Process(ctx, req)
	}

	return ExecCommand(logId, req.Timeout, fn)
}
