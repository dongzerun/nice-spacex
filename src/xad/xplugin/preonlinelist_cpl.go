package xplugin

import (
	"golang.org/x/net/context"

	"common"
	"defined"
	"xuser"
)

var _ Matcher = (*PreOnlineList)(nil)

// 临时上线：20161013 一天后即不在用
var (
	TmpRecommandUids  = []int64{13023610, 21566023, 21475286}
	TmpRecommandAdIds = []int{1037, 1038, 1039, 1040, 1041, 1042, 1043}
)

// 标签匹配
type PreOnlineList struct {
	BaseMatch
}

// 对于白名单用户才有此条件
func (pll *PreOnlineList) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	exists := xuser.IsInPreOnlineList(ui.Uid)
	// common.Infof("logid:%v adstatus:%s,uid:%d whitelist:%v", ctx.Value("logid"), pll.GetAdStatus(), ui.Uid, exists)
	if pll.GetAdStatus() == "test" && ui.Uid != 0 && exists { // 预上线广告
		return true
	}

	if ui.Req != nil {
		if defined.IsDeviceInPreOnline(ui.Req.DeviceID) {
			return true
		}
	}

	// 临时上线：20161013 一天后即不在用
	ad := pll.GetAdId()
	uid := ui.Uid
	for _, id := range TmpRecommandAdIds {
		if ad == id {
			for _, u := range TmpRecommandUids {
				if uid == u {
					common.Infof("logid:%v adid:%d uid:%d hit tmp hack20161013", ctx.Value("logid"), ad, uid)
					return true
				}
			}
		}
	}

	return false
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "preonline_list_enum"
	RegisterMatcher(name, func() Matcher {
		pll := &PreOnlineList{}
		pll.BaseMatch.PluginName = name
		return pll
	})
}
