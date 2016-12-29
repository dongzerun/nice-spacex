package xplugin

import (
	"golang.org/x/net/context"

	"xuser"
)

var _ Matcher = (*ExcludeRegionEnum)(nil)

// 渠道匹配
type ExcludeRegionEnum struct {
	BaseMatch
	Regions string
}

func (ere *ExcludeRegionEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ere.IsAll() {
		return true
	}

	return true
}

func init() {
	name := "exclude_region_enum"
	RegisterMatcher(name, func() Matcher {
		ire := &ExcludeRegionEnum{}
		ire.BaseMatch.PluginName = name
		return ire
	})
}
