package xplugin

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	"xuser"
	"xutil/cache"
	database "xutil/db"
	// "xutil/hack"
)

var (
	_ Matcher = (*IdMatchEnum)(nil)
)

// "10849_exists":{"sids":[158887372],"sid_count":1}
// 在 show_cache 中用户发过标签的缓存格式，sids 可能是数组，但是Others就是map
// 格式不统一，很讨厌
type TagDetail struct {
	Sids     interface{} `json:"sids"`
	SidCount int         `json:"sid_count"`
}

// 新增用户匹配
type IdMatchEnum struct {
	BaseMatch
	Ids []int64
}

func (ime *IdMatchEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {

	if ime.IsAll() {
		// 默认是所有人可见，那么返回 true
		return true
	}

	select {
	case <-ctx.Done():
		common.Warningf("logid:%v IdMatchEnum timeout or canceled, directly return", ctx.Value("logid"))
		return false
	default:
	}

	switch ime.GetName() {
	// 发送过自定义标签
	case "public_undefined_tag_enum":
		for _, tagid := range ime.Ids {
			ok, err := IsPublicTagInCache(ctx, ui.Uid, tagid)
			if err != nil {
				common.Warningf("logid:%v IdMatchEnum IsPublicTag uid:%d, tagid:%d err:%s ", ctx.Value("logid"), ui.Uid, tagid, err.Error())
				continue
			}
			if ok {
				return true
			}
		}
		return false
	// 发送过地点标签
	case "public_point_tag_enum":
		for _, tagid := range ime.Ids {
			ok, err := IsPublicTag(ui.Uid, tagid)
			if err != nil {
				common.Warningf("logid:%v IdMatchEnum public_point_tag_enum uid:%d, tagid:%d err:%s ", ctx.Value("logid"), ui.Uid, tagid, err.Error())
				continue
			}
			if ok {
				return true
			}
		}
		return false
	// 关注过自定义标签,关注过地点标签
	case "follow_undefined_tag_enum", "follow_point_tag_enum":
		for _, tagid := range ime.Ids {
			ok, err := IsFollowTag(ui.Uid, tagid)
			if err != nil {
				common.Warningf("logid:%v IdMatchEnum IsFollowTag uid:%d, tagid:%d err:%s ", ctx.Value("logid"), ui.Uid, tagid, err.Error())
				continue
			}
			if ok {
				return true
			}
		}
		return false
	// 使用过特定贴纸包
	case "use_package_id_enum":
		return true
	// 使用过特定贴纸
	case "use_paster_id_enum":
		for _, pasterid := range ime.Ids {
			ok, err := IsUsePaster(ui.Uid, pasterid)
			if err != nil {
				common.Warningf("logid:%v IdMatchEnum IsUsePaster uid:%d, pasterid:%d err:%s ", ctx.Value("logid"), ui.Uid, pasterid, err.Error())
				continue
			}
			if ok {
				return true
			}
		}
		return false
	}
	common.Warningf("logid:%v uid:%d IdMatchEnum unknow name ", ctx.Value("logid"), ui.Uid)
	return false
}

func init() {
	names := []string{"public_undefined_tag_enum", "public_point_tag_enum",
		"follow_undefined_tag_enum", "follow_point_tag_enum", "use_package_id_enum",
		"use_paster_id_enum"}

	for _, name := range names {
		matcherName := name
		RegisterMatcher(name, func() Matcher {
			m := &IdMatchEnum{}
			m.BaseMatch.PluginName = matcherName
			return m
		})
	}
}

func GenPublicTagSql(tagid int64) string {
	return fmt.Sprintf("SELECT COUNT(*) from kk_user_show_image_tag_idx_%d where uid=? and tagid=?", tagid%1000)
}

// 是否使用过某些tag
// 无论是官方还是非官方
func IsPublicTag(uid int64, tagid int64) (bool, error) {
	db := database.UserTag.P.GetConn()
	if db == nil {
		return false, common.MySQLPoolTimeOutError
	}
	defer database.UserTag.P.Release(db)

	sql := GenPublicTagSql(tagid)
	common.Infof("IsPublicTag uid:%d tagid:%d sql:%s", uid, tagid, sql)

	rows, err := db.Query(sql, uid, tagid)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// 查看 count(*) 这样其实不太好，当前能用就行
	for rows.Next() {
		var count int
		e := rows.Scan(&count)
		if e != nil {
			return false, e
		}
		return count >= 1, nil
	}
	return false, nil
}

func IsPublicTagInCache(ctx context.Context, uid int64, tagid int64) (bool, error) {
	key := fmt.Sprintf(defined.KeyTagUserCacheFormat, uid)

	data, err := cache.GlobalShowCache.GetByte(key)
	if err != nil {
		return false, err
	}

	tag := make(map[string]*TagDetail)

	err = json.Unmarshal(data, &tag)
	if err != nil {
		return false, err
	}

	// common.Infof("logid:%v IsPublicTagInCache uid:%d, tagdata:%s", ctx.Value("logid"), uid, hack.String(data))

	key = fmt.Sprintf(defined.KeyTagPublicTagFormat, tagid)

	tagDetail, exists := tag[key]
	if !exists {
		// 不存在这个标签，没有打过
		// common.Infof("logid:%v IsPublicTagInCache uid:%d, tagid:%d key:%s not exists", ctx.Value("logid"), uid, tagid, key)
		return false, nil
	}

	return tagDetail.SidCount >= 1, nil
}

func GenUsePasterSql(pasterid int64) string {
	return fmt.Sprintf("select count(*) from user_show_paster_%d where pid=? and uid=?", pasterid%1000)
}

// 用户是否使用过某贴纸
// 当前走数据库，可能开销会很大
func IsUsePaster(uid int64, pasterid int64) (bool, error) {
	db := database.UserPaster.P.GetConn()
	if db == nil {
		return false, common.MySQLPoolTimeOutError
	}
	defer database.UserPaster.P.Release(db)

	sql := GenUsePasterSql(pasterid)
	common.Infof("IsUsePaster uid:%d pasterid:%d sql:%s", uid, pasterid, sql)

	rows, err := db.Query(sql, uid, pasterid)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// 查看 count(*) 这样其实不太好，当前能用就行
	for rows.Next() {
		var count int
		e := rows.Scan(&count)
		if e != nil {
			return false, e
		}
		return count >= 1, nil
	}
	return false, nil
}

// 某个用户是否使用过某贴纸包
// 优先级不高，暂时先不实现，数据库不好处理
func IsUsePackage(uid int64, pasterid int64) (bool, error) {
	return true, nil
}

// 查看某人是否关注过特定类型的某个标签
// brand 品牌
// undefined 未定义
// point 地点
// custom_point 自定义地点 UserAsset
func IsFollowTag(uid int64, tagid int64) (bool, error) {
	db := database.UserAsset.P.GetConn()
	if db == nil {
		return false, common.MySQLPoolTimeOutError
	}
	defer database.UserAsset.P.Release(db)

	common.Infof("IsFollowTag uid:%d tagid:%d sql:%s", uid, tagid, defined.GetUserAssetCountSql)

	rows, err := db.Query(defined.GetUserAssetCountSql, uid, tagid)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// 查看 count(*) 这样其实不太好，当前能用就行
	for rows.Next() {
		var count int
		e := rows.Scan(&count)
		if e != nil {
			return false, e
		}
		return count >= 1, nil
	}
	return false, nil
}
