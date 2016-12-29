package xplugin

import (
	"time"

	"golang.org/x/net/context"

	// "common"
	"xuser"
)

var _ Matcher = (*ValidTimeRange)(nil)

// 渠道匹配
type ValidTimeRange struct {
	BaseMatch
	Time []int64
}

func (vtr *ValidTimeRange) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if vtr.IsAll() {
		return true
	}

	now := time.Now().Unix()
	// common.Debugf("logid:%v Valid time range:%d", ctx.Value("logid"), vtr.Time)
	if len(vtr.Time) != 2 {
		return false
	}

	return vtr.Time[0] <= now && vtr.Time[1] >= now
}

func (vtr *ValidTimeRange) Remaining(ctx context.Context) int {

	now := time.Now().Unix()

	if len(vtr.Time) != 2 {
		return 0
	}

	remaining := vtr.Time[1] - now
	// common.Debugf("logid:%v Valid time remaining:%d", ctx.Value("logid"), remaining)
	return int(remaining)
}

type Range struct {
	Start int64 `json:"start"`
	End   int64 `json:"finish"`
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "valid_time_between"
	RegisterMatcher(name, func() Matcher {
		vtr := &ValidTimeRange{}
		vtr.BaseMatch.PluginName = name
		return vtr
	})
}
