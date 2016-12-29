package xplugin

import (
	"golang.org/x/net/context"

	"xuser"
)

var (
	_ Matcher = (*VisibleUsersEnum)(nil)
)

// 可见用户匹配
type VisibleUsersEnum struct {
	BaseMatch
	Users []int64
}

func (vue *VisibleUsersEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {

	if vue.IsAll() {
		// 默认是所有人可见，那么返回 true
		return true
	}

	return SimpleInt64Match(vue.Users, ui.Uid)
}

func init() {
	name := "visible_users_enum"
	RegisterMatcher(name, func() Matcher {
		vue := &VisibleUsersEnum{}
		vue.BaseMatch.PluginName = name
		return vue
	})
}
