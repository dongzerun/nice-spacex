package xplugin

import (
	"golang.org/x/net/context"

	"common"
	"xuser"
)

var _ Matcher = (*UserGroupEnum)(nil)

// 尾号匹配，取用户ID最后一位和Groups进行匹配
type UserGroupEnum struct {
	BaseMatch
	Groups []int
}

func (uge *UserGroupEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if uge.IsAll() {
		return true
	}

	for _, u := range uge.GetWL() {
		if ui.Uid == int64(u) {
			// 命中白名单用户
			common.Infof("logid:%v uid:%d group enum hit ad inner white list", ctx.Value("logid"), ui.Uid)
			return true
		}
	}

	return SimpleIntMatch(uge.Groups, int(ui.Uid%10))
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
// "user_group"
func init() {
	name := "user_group_enum"
	RegisterMatcher(name, func() Matcher {
		uge := &UserGroupEnum{}
		uge.BaseMatch.PluginName = name
		return uge
	})
}
