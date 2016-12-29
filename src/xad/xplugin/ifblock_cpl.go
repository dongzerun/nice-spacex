package xplugin

import (
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	"xuser"
	"xutil/cache"
)

var (
	_ Matcher = (*IfBlock)(nil)
)

// 新增用户匹配
type IfBlock struct {
	BaseMatch
}

// 如果blocked return false
// 如果未被blocked return true
func (ib *IfBlock) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v IfBlock timeout or canceled, directly return", ctx.Value("logid"))
		return false
	default:
	}

	key := fmt.Sprintf(defined.KeyBlockActionFormat, ui.Uid, ib.GetAdId())

	exists, _ := cache.GlobalRedisCache.Exists(key)
	// common.Infof("logid:%v uid:%d key:%s exists:%v", ctx.Value("logid"), ui.Uid, key, exists)
	if exists {
		return false
	}
	return true
}

func init() {
	name := "if_block"
	RegisterMatcher(name, func() Matcher {
		ib := &IfBlock{}
		ib.BaseMatch.PluginName = name
		return ib
	})
}
