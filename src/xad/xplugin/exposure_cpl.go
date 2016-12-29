package xplugin

import (
	"golang.org/x/net/context"

	"common"
	"xuser"
	"xutil/limit"
)

var _ Matcher = (*ExposureRatioEnum)(nil)

// 对广告曝光量进行限制, 达到一定比例会拒绝展示
// ratio ~ (1,10]
// 一天分成 2 * 24 * 60 = 2880 个时间片段，每片 30s
// 那么曝光比例就是 N 段时间内，只有1段给予曝光，其它 N-1 段时间不展示
// 比较注意的点在于： 这N段时间内，能展示的时间段是随机的，不能都固定在某一个
// 以防止多个广告同时抢占一个时间片展示，而造成的实际展示量不足
type ExposureRatioEnum struct {
	BaseMatch
	Limit *limit.LimitServer
}

func (ere *ExposureRatioEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ere.IsAll() {
		return true
	}

	exposure := ere.Limit.Exposure()
	common.Warningf("logid:%v Exposure ad:%d adstatus:%s exposure:%v", ctx.Value("logid"), ere.GetAdId(), ere.GetAdStatus(), exposure)
	return exposure
}

func init() {
	name := "exposure_enum"
	RegisterMatcher(name, func() Matcher {
		ere := &ExposureRatioEnum{}
		ere.BaseMatch.PluginName = name
		return ere
	})
}
