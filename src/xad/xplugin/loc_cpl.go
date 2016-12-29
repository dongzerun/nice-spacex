package xplugin

import (
	"strings"

	"golang.org/x/net/context"

	"common"
	"xuser"
	"xutil/loc"
)

var _ Matcher = (*ProvinceEnum)(nil)
var _ Matcher = (*CityEnum)(nil)

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "city_enum"
	RegisterMatcher(name, func() Matcher {
		pe := &CityEnum{}
		pe.BaseMatch.PluginName = name
		return pe
	})
}

// 省市匹配
type ProvinceEnum struct {
	BaseMatch
	Provs []string
}

// 市区匹配
type CityEnum struct {
	BaseMatch
	Cities []string
}

func (ce *CityEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ce.IsAll() {
		return true
	}
	loc := getLoc(ui, "city")
	common.Infof("logid:%v CityEnum get loc %s", ctx.Value("logid"), loc)
	for _, c := range ce.Cities {
		if strings.HasPrefix(loc, c) {
			common.Infof("logid:%v CityEnum get loc %s match %s", ctx.Value("logid"), loc, c)
			return true
		}
	}
	return false
}

func (pe *ProvinceEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if pe.IsAll() {
		return true
	}
	loc := getLoc(ui, "province")
	common.Infof("logid:%v ProvinceEnum get loc %s", ctx.Value("logid"), loc)
	for _, c := range pe.Provs {
		if strings.HasPrefix(loc, c) {
			common.Infof("logid:%v ProvinceEnum get loc %s match %s", ctx.Value("logid"), loc, c)
			return true
		}
	}
	return false
}

func getLoc(ui *xuser.UserInfo, choice string) string {
	return loc.GetLoc(*ui.Req.Extra.Longitude, *ui.Req.Extra.Latitude, choice)
}
