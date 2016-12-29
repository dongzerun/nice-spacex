package xplugin

import (
	"golang.org/x/net/context"

	"xuser"
)

var _ Matcher = (*GenderEnum)(nil)

// 性别匹配
type GenderEnum struct {
	BaseMatch
	Genders []string
}

func (ge *GenderEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ge.IsAll() {
		return true
	}
	//return xutil.SimpleStrMatch(ge.Genders, ui)
	return SimpleStrMatch(ge.Genders, []string{ui.Gender})
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "gender_enum"
	RegisterMatcher(name, func() Matcher {
		ge := &GenderEnum{}
		ge.BaseMatch.PluginName = name
		return ge
	})
}
