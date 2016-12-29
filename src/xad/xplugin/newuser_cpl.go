package xplugin

import (
	"time"

	"golang.org/x/net/context"

	"xuser"
)

var (
	_ Matcher = (*NewUserEnum)(nil)
)

// 新增用户匹配
type NewUserEnum struct {
	BaseMatch
	Day int
}

func (nue *NewUserEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {

	if nue.IsAll() {
		// 默认是所有人可见，那么返回 true
		return true
	}

	return int(time.Now().UnixNano())-ui.CreateTime <= nue.Day*86400
}

func init() {
	name := "is_new_enum"
	RegisterMatcher(name, func() Matcher {
		nue := &NewUserEnum{}
		nue.BaseMatch.PluginName = name
		return nue
	})
}
