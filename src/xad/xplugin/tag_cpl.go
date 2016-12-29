package xplugin

import (
	"golang.org/x/net/context"

	//"common"
	//"strings"
	"xuser"
)

var _ Matcher = (*TagListEnum)(nil)

// 标签匹配
type TagListEnum struct {
	BaseMatch
	TagClass []string
}

func (tle *TagListEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	//2016-12-25紧急上线
	return true
	//if tle.IsAll() {
	//	return true
	//}
	//for ui_tag_class_index, _ := range ui.TagClass {
	//	for tle_tag_class_index, _ := range tle.TagClass {
	//		if strings.EqualFold(
	//			ui.TagClass[ui_tag_class_index].ClassCn,
	//			tle.TagClass[tle_tag_class_index]) {
	//			common.Infof("logid:%v matched ad:%v tag class:%v user class:%v", ctx.Value("logid"), tle.AdId, tle.TagClass[tle_tag_class_index], ui.TagClass[ui_tag_class_index].ClassCn)
	//			return true
	//		}
	//	}
	//}
	//return false
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "tag_list_enum"
	RegisterMatcher(name, func() Matcher {
		tle := &TagListEnum{}
		tle.BaseMatch.PluginName = name
		return tle
	})
}
