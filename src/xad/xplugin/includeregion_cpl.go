package xplugin

import (
	"golang.org/x/net/context"

	"xuser"
)

var _ Matcher = (*IncludeRegionEnum)(nil)

// 渠道匹配
type IncludeRegionEnum struct {
	BaseMatch
	Regions string
}

func (ire *IncludeRegionEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ire.IsAll() {
		return true
	}

	return true
}

func init() {
	name := "include_region_enum"
	RegisterMatcher(name, func() Matcher {
		ire := &IncludeRegionEnum{}
		ire.BaseMatch.PluginName = name
		return ire
	})
}
