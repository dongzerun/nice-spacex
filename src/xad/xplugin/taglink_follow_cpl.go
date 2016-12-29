package xplugin

import (
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	"xuser"
	"xutil/cache"
)

var _ Matcher = (*TaglinkFollowEnum)(nil)

// 按钮标签外连专用的过滤器，只有请求用户关注了 followuid 才为真
type TaglinkFollowEnum struct {
	BaseMatch
	FollowUid int64
}

func (tfe *TaglinkFollowEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	return IsTagLinkFollow(ctx, ui.Uid, tfe.FollowUid)
}

func IsTagLinkFollow(ctx context.Context, uid, follow int64) bool {
	key := fmt.Sprintf(defined.KeyUserFollowingFormat, uid)
	ts, _ := cache.GlobalSocialCache.ScoreInt64(key, follow)

	common.Warningf("logid:%v ScoreInt64:%s member:%d score:%d", ctx.Value("logid"), key, follow, ts)
	return ts > 0
}

func init() {
	name := "taglink_follow_enum"
	RegisterMatcher(name, func() Matcher {
		tfe := &TaglinkFollowEnum{}
		tfe.BaseMatch.PluginName = name
		return tfe
	})
}
