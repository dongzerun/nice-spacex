package xplugin

import (
	"golang.org/x/net/context"

	"common"
	"defined"
	"xuser"
	database "xutil/db"
)

var (
	_ Matcher = (*FollowUsersEnum)(nil)
)

// 新增用户匹配
type FollowUsersEnum struct {
	BaseMatch
	Users []int64
}

func (fue *FollowUsersEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {

	if fue.IsAll() {
		// 默认是所有人可见，那么返回 true
		return true
	}

	select {
	case <-ctx.Done():
		common.Warningf("logid:%v FollowUsesEnum timeout or canceled, directly return", ctx.Value("logid"))
		return false
	default:
	}

	for _, u := range fue.Users {
		follow, err := IsFollow(ui.Uid, u)
		if err != nil {
			common.Warningf("logid:%v Filter FollowUsers IsFollow err:%s", ctx.Value("logid"), err.Error())
		}

		if follow {
			common.Infof("logid:%v Filter FollowUsers match success uid:%d, following:%d", ctx.Value("logid"), ui.Uid, u)
			return true
		}
	}
	return false
}

// 从数据库里查询是否关注过某人
func IsFollow(uid int64, follow int64) (bool, error) {
	db := database.CoreShow.P.GetConn()
	if db == nil {
		return false, common.MySQLPoolTimeOutError
	}
	defer database.CoreShow.P.Release(db)

	rows, err := db.Query(defined.GetUserFollowSql, uid, follow)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// getUserFollowSql limit 1 确保永远只有一条数据
	for rows.Next() {
		var exists int
		e := rows.Scan(&exists)
		if e != nil {
			return false, e
		}
		return exists == 1, nil
	}
	return false, nil
}

func init() {
	name := "follow_users_enum"
	RegisterMatcher(name, func() Matcher {
		fue := &FollowUsersEnum{}
		fue.BaseMatch.PluginName = name
		return fue
	})
}
