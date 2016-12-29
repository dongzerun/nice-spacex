package xplugin

import (
	"golang.org/x/net/context"

	"strings"
	"xuser"
)

var (
	_ Matcher = (*IosVersionEnum)(nil)
	_ Matcher = (*AndroidVersionEnum)(nil)
	_ Matcher = (*IosVersionRange)(nil)
	_ Matcher = (*AndroidVersionRange)(nil)
)

// IOS版本匹配
type IosVersionEnum struct {
	BaseMatch
	Versions [][]int
}

func (ive *IosVersionEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {

	if ive.IsAll() || strings.ToLower(*ui.Req.Extra.DeviceOs) == "android" {
		// 条件是：
		// 1: 如果是苹果所有版本 则返回true
		// 2: 如果用户是android用户，则无需匹配ios的条件
		return true
	}

	if ive.IsFalse() {
		// 如果广告上指定不是匹配 ios的，则返回 false
		return false
	}

	if ui.Req != nil && ui.Req.Extra != nil && ui.Req.Extra.AppVersion != nil {
		v, err := IntSlice(*ui.Req.Extra.AppVersion)
		if err != nil {
			return false
		}
		return VersionEnumMatch(ive.Versions, v)
	}
	return false
}

type IosVersionRange struct {
	BaseMatch
	Versions [][]int
}

func (ivr *IosVersionRange) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ivr.IsAll() || strings.ToLower(*ui.Req.Extra.DeviceOs) == "android" {
		// 条件是：
		// 1: 如果是苹果所有版本 则返回true
		// 2: 如果用户是android用户，则无需匹配ios的条件
		return true
	}
	if ivr.IsFalse() {
		// 如果广告上指定不是匹配 ios的，则返回 false
		return false
	}
	if ui.Req != nil && ui.Req.Extra != nil && ui.Req.Extra.AppVersion != nil {
		v, err := IntSlice(*ui.Req.Extra.AppVersion)
		if err != nil {
			return false
		}
		return VersionRangeMatch(ivr.Versions, v)
	}
	return false
}

// 安卓版本匹配
type AndroidVersionEnum struct {
	BaseMatch
	Versions [][]int
}

func (ave *AndroidVersionEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if ave.IsAll() || strings.ToLower(*ui.Req.Extra.DeviceOs) == "ios" {
		// 条件是：
		// 1: 如果是android所有版本 则返回true
		// 2: 如果用户是ios用户，则无需匹配android的条件
		return true
	}
	if ave.IsFalse() {
		// 如果广告上指定不是匹配 android的，则返回 false
		return false
	}
	if ui.Req != nil && ui.Req.Extra != nil && ui.Req.Extra.AppVersion != nil {
		v, err := IntSlice(*ui.Req.Extra.AppVersion)
		if err != nil {
			return false
		}
		return VersionEnumMatch(ave.Versions, v)
	}
	return false
}

type AndroidVersionRange struct {
	BaseMatch
	Versions [][]int
}

func (avr *AndroidVersionRange) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if avr.IsAll() || strings.ToLower(*ui.Req.Extra.DeviceOs) == "ios" {
		// 条件是：
		// 1: 如果是android所有版本 则返回true
		// 2: 如果用户是ios用户，则无需匹配android的条件
		return true
	}
	if avr.IsFalse() {
		// 如果广告上指定不是匹配 android的，则返回 false
		return false
	}
	if ui.Req != nil && ui.Req.Extra != nil && ui.Req.Extra.AppVersion != nil {
		v, err := IntSlice(*ui.Req.Extra.AppVersion)
		if err != nil {
			return false
		}
		return VersionRangeMatch(avr.Versions, v)
	}
	return false
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "ios_app_version_enum"
	RegisterMatcher(name, func() Matcher {
		ive := &IosVersionEnum{}
		ive.BaseMatch.PluginName = name
		return ive
	})

	name = "ios_app_version_min"
	RegisterMatcher(name, func() Matcher {
		ivr := &IosVersionRange{}
		ivr.BaseMatch.PluginName = name
		return ivr
	})

	name = "android_app_version_enum"
	RegisterMatcher(name, func() Matcher {
		ave := &AndroidVersionEnum{}
		ave.BaseMatch.PluginName = name
		return ave
	})

	name = "android_app_version_min"
	RegisterMatcher(name, func() Matcher {
		avr := &AndroidVersionRange{}
		avr.BaseMatch.PluginName = name
		return avr
	})
}
