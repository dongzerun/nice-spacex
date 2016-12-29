package xad

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	simple "github.com/bitly/go-simplejson"

	"common"
	"defined"
	"xad/xplugin"
	"xutil/hack"
	"xutil/limit"
)

type Filter struct {
	FId             int    `json:"fid"`   // 条件ID
	AId             int    `json:"aid"`   // 对应的广告ID
	Name            string `json:"fname"` // 条件名称 enum("valid_time", "max_display_num", "ios_app_version","android_app_version", "app_channel","gender","city", "tag_list")
	Type            string `json:"type"`  // 运算 enum("enum", "min", "between")
	Value           string `json:"value"` // 取值
	CreateTime      int    `json:"ctime"` // 创建时间
	UpdateTime      int    `json:"utime"` // 更新时间
	xplugin.Matcher        // Filter 所对应的匹配模块
}

func (f *Filter) GenerateMatcher(ad *Ad) error {
	m, err := xplugin.GetMatcher(f.Name, f.Type)
	if err != nil {
		return err
	}

	m.SetAdId(ad.Id)
	m.SetAdStatus(ad.Status)
	m.SetAdArea(ad.Area)
	f.Matcher = m
	return nil
}

// return filter pointer
func fillPreOnline(ad *Ad) (*Filter, error) {
	f := new(Filter)
	f.Name = "preonline_list"
	f.Type = "enum"

	err := f.GenerateMatcher(ad)
	return f, err
}

// return error ,and modify f *Filter if possible
func fillValidTime(f *Filter, ad *Ad) error {
	if f.Value == "all" {
		f.Matcher.SetAll()
		return nil
	}

	rge := &xplugin.Range{}
	err := json.Unmarshal(hack.Slice(f.Value), rge)
	if err != nil {
		common.Warning("updateFilter range Unmarshal failed ", err.Error(), f.Value)
		return err
	}

	vtr, ok := f.Matcher.(*xplugin.ValidTimeRange)
	if !ok {
		common.Warning("updateFilter range assert vtr failed ")
		return defined.ErrMatcherAssert
	}
	vtr.Time = []int64{rge.Start, rge.End}

	return nil
}

// return error ,and modify f *Filter if possible
func fillVisibleUsers(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	vue, ok := f.Matcher.(*xplugin.VisibleUsersEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}

	vue.Users = make([]int64, 0)

	for _, s := range vs {
		u, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			common.Warning("updateFilter visble_users_enum parse user failed ", s)
			continue
		}
		vue.Users = append(vue.Users, u)
	}
	return nil
}

func fillFollowUsers(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	fue, ok := f.Matcher.(*xplugin.FollowUsersEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}

	fue.Users = make([]int64, 0)

	for _, s := range vs {
		u, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			common.Warning("updateFilter visble_users_enum parse user failed ", s)
			continue
		}
		fue.Users = append(fue.Users, u)
	}
	return nil
}

func fillIdMatch(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	ime, ok := f.Matcher.(*xplugin.IdMatchEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}

	ime.Ids = make([]int64, 0)

	for _, s := range vs {
		id, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			common.Warning("updateFilter fillIdMatch parse user failed ", s)
			continue
		}
		ime.Ids = append(ime.Ids, id)
	}

	return nil
}

func fillIsNew(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	if len(vs) != 1 {
		return defined.ErrIsNewNum
	}

	d, err := strconv.Atoi(vs[0])
	if err != nil {
		return err
	}

	nue, ok := f.Matcher.(*xplugin.NewUserEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	nue.Day = d

	return nil
}

func fillMaxDisplayNum(f *Filter, ad *Ad) error {
	if f.Value == "all" {
		f.Matcher.SetAll()
		return nil
	}
	num, err := strconv.Atoi(f.Value)
	if err != nil {
		return err
	}

	mdm, ok := f.Matcher.(*xplugin.DsplyNumEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	mdm.Num = num
	return nil
}

func fillIosVersion(f *Filter, ad *Ad) error {
	// if f.Value == "all" {
	// 	m.SetAll()
	// 	continue
	// }
	// // min         | 3.9.0
	// // enum        | 3.9.1
	// vs := strings.Split(f.Value, ",")
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	for i, _ := range vs {
		if vs[i] == "none" {
			f.Matcher.SetFalse()
			return nil
		}
	}

	switch f.Type {
	case "min":
		if len(vs) != 1 {
			common.Warning("updateFilter vs length expected 1 now ", len(vs))
		}
		v, err := IntSlice(vs[0])
		if err != nil {
			return err
		}

		ivr, ok := f.Matcher.(*xplugin.IosVersionRange)
		if !ok {
			return defined.ErrMatcherAssert
		}
		ivr.Versions = [][]int{v, MaxVersion}

	case "enum":
		ive, ok := f.Matcher.(*xplugin.IosVersionEnum)
		if !ok {
			return defined.ErrMatcherAssert
		}

		ive.Versions = make([][]int, 0, 1)
		for _, data := range vs {
			v, err := IntSlice(data)
			if err != nil {
				common.Warning("updateFilter vs IntSlice failed ", err.Error(), data)
				continue
			}
			ive.Versions = append(ive.Versions, v)
		}
	}

	return nil
}

func fillAndroidVersion(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	for i, _ := range vs {
		if vs[i] == "none" {
			f.Matcher.SetFalse()
			return nil
		}
	}

	switch f.Type {
	case "min":
		if len(vs) != 1 {
			common.Warning("updateFilter vs length expected 1 now ", len(vs))
		}
		v, err := IntSlice(vs[0])
		if err != nil {
			return err
		}

		avr, ok := f.Matcher.(*xplugin.AndroidVersionRange)
		if !ok {
			return defined.ErrMatcherAssert
		}
		avr.Versions = [][]int{v, MaxVersion}

	case "enum":
		ave, ok := f.Matcher.(*xplugin.AndroidVersionEnum)
		if !ok {
			return defined.ErrMatcherAssert
		}
		ave.Versions = make([][]int, 0, 1)
		for _, data := range vs {
			v, err := IntSlice(data)
			if err != nil {
				common.Warning("updateFilter vs IntSlice failed ", err.Error(), data)
				continue
			}
			ave.Versions = append(ave.Versions, v)
		}
	}

	return nil
}

func fillGender(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	ge, ok := f.Matcher.(*xplugin.GenderEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	ge.Genders = vs

	return nil
}

func fillCity(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	pe, ok := f.Matcher.(*xplugin.CityEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	pe.Cities = vs

	return nil
}

func fillTagList(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	tle, ok := f.Matcher.(*xplugin.TagListEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	tle.TagClass = vs

	return nil
}

func fillUserGroup(f *Filter, ad *Ad) error {
	isAll, vs, e := AnalyzValue(f.Value)
	if e != nil {
		return defined.ErrAnalyzValue
	}

	if isAll {
		f.Matcher.SetAll()
		return nil
	}

	uge, ok := f.Matcher.(*xplugin.UserGroupEnum)
	if !ok {
		return defined.ErrMatcherAssert
	}
	ivs := make([]int, 0, 1)
	for _, str := range vs {
		n, err := strconv.Atoi(str)
		if err == nil {
			ivs = append(ivs, n)
		}
	}

	uge.Groups = ivs
	return nil
}

func fillWhiteList(f *Filter, ad *Ad) error {
	_, vslice, err := AnalyzValue(f.Value)
	if err != nil {
		return defined.ErrAnalyzValue
	}

	wl := make([]int, 0)
	for _, s := range vslice {
		uid, err := strconv.Atoi(s)
		if err == nil {
			common.Infof("updateFilter white_list add uid:%d to white_list", uid)
			wl = append(wl, uid)
			continue
		}
		common.Warningf("updateFilter white_list strconv %s err:%s", s, err.Error())
	}

	ad.WhiteList = wl
	return nil
}

// updateFilter 的升级版本，将大函数打散，易于维护
func updateFilter2(ad *Ad) {
	fs, err := GetFilterByAdID(ad.Id)
	if err != nil {
		common.Warning("updateFilter GetFilterByAdID err ", err.Error())
		return
	}

	if ad.Filters == nil {
		ad.Filters = make([]*Filter, 0, 2)
	}
	// 只有预上线的广告才有此filter
	// 预上线的广告只针对 preonline list 里的人开放
	if ad.Status == "test" {
		f, err := fillPreOnline(ad)
		if err != nil {
			common.Warningf("updateFilter xpreonline err:%s", err.Error())
		} else {
			ad.Filters = append(ad.Filters, f)
		}
	}

	AddTagLinkFollowCpl(ad)

	for _, f := range fs {
		common.Debug("update ad filter ", ad.Id, f)

		if f.Name == "white_list" {
			fillWhiteList(f, ad)
			continue
		}
		var err error

		// 根据 f.Name f.Type 自动生成 Matcher
		err = f.GenerateMatcher(ad)
		if err != nil {
			common.Warningf("updateFilter2 FillMatcher err:%s name:%s type:%s", err.Error(), f.Name, f.Type)
			continue
		}

		// 解析value 填充 Matcher
		switch f.Name {
		case "valid_time":
			err = fillValidTime(f, ad)
		case "visible_users":
			err = fillVisibleUsers(f, ad)
		case "follow_users":
			err = fillFollowUsers(f, ad)
		case "public_undefined_tag", "public_point_tag",
			"follow_undefined_tag", "follow_point_tag", "use_package_id",
			"use_paster_id":
			err = fillIdMatch(f, ad)
		case "is_new":
			err = fillIsNew(f, ad)
		case "max_display_num":
			err = fillMaxDisplayNum(f, ad)
		case "ios_app_version":
			err = fillIosVersion(f, ad)
		case "android_app_version":
			err = fillAndroidVersion(f, ad)
		case "gender":
			err = fillGender(f, ad)
		case "city":
			err = fillCity(f, ad)
		case "include_region":
			err = defined.ErrUnImplementedPlugin
		case "exclude_region":
			err = defined.ErrUnImplementedPlugin
		case "tag_list":
			err = fillTagList(f, ad)
		case "user_group":
			err = fillUserGroup(f, ad)
		case "white_list":
			err = fillWhiteList(f, ad)
		default:
			common.Warningf("ad:%d, get an not being used filter: %s", ad.Id, f.Name)
		}

		if err == nil {
			ad.Filters = append(ad.Filters, f)
		} else {
			common.Warningf("updateFilter2 name:%s type:%s err:%s value:%s", f.Name, f.Type, err.Error(), f.Value)
		}
	}

	// 广告有一个附属的白名单用户
	// 名单内的用户不受 max_display_num 控制，无限弹出
	for _, f := range ad.Filters {
		if f.Matcher.GetName() == "max_display_num_enum" || f.Matcher.GetName() == "user_group_enum" {
			f.Matcher.SetWL(ad.WhiteList)
			common.Infof("updateFilter max_display_num|user_group_enum after add whitelist %v", f.Matcher.GetWL())
		}
	}

	// 点击反馈：
	// 用户可以屏蔽该广告
	// 默认最后都要加一个 if block 过滤条件
	ibf := &Filter{
		AId:     ad.Id,
		Name:    "if_block",
		Matcher: new(xplugin.IfBlock),
	}

	ibf.Matcher.SetAdId(ad.Id)

	ad.Filters = append(ad.Filters, ibf)

	UpdateExposureEnum(ad)

	// 编排过滤条件
	ad.arrange()
}

func UpdateExposureEnum(ad *Ad) {
	// 增加 ExposureRatioEnum 过滤条件，由于要用到valid_time中的start 所以只能单独写在这里
	// 如果 exposure_enum更新失败，那么存在的filter要设置SetAll, 这样才不影响结果
	for idx, f := range ad.Filters {
		if f.Matcher.GetName() != "exposure_enum" {
			continue
		}

		// filter 生成 mathcer
		err := f.GenerateMatcher(ad)
		if err == nil {
			ere, ok := f.Matcher.(*xplugin.ExposureRatioEnum)
			if !ok {
				common.Warning("AddExposureEnum exposure ratio enum assert failed ")
				return
			}

			ratio := strings.Split(f.Value, "/")
			if len(ratio) != 2 {
				common.Warningf("AddExposureEnum filter value must like 3/7 but get %s", f.Value)
				ere.SetAll()
				return
			}

			valid, _ := strconv.Atoi(ratio[0]) // 分子，有效桶数
			num, _ := strconv.Atoi(ratio[1])   // 分母，总桶数

			if num == 0 || valid == 0 || (num <= valid) {
				common.Warningf("AddExposureEnum ad:%d num:%d valid:%d unavilable", f.AId, num, valid)
				ere.SetAll()
				return
			}

			//需要找到 ad start 时间，所以要遍历ad.Filter找到 "valid_time_between"
			var (
				vtb   *xplugin.ValidTimeRange
				found bool
			)

			for _, fl := range ad.Filters {
				if fl.Matcher.GetName() == "valid_time_between" {
					vtb, found = fl.Matcher.(*xplugin.ValidTimeRange)
					break
				}
			}

			if vtb == nil || !found {
				common.Warningf("AddExposureEnum valid time range not found")
				ere.SetAll()
				return
			}

			// 这里假设 vtb 一定要准备好，并且 vtb.Time 一定是两个元素的slice，分别为广告的起止时间 unix timestamp
			ere.Limit = limit.NewLimitServer(vtb.Time[0], num, valid, fmt.Sprintf("ad:%s", f.AId))
			ere.SetWL(ad.WhiteList)
			return
		}

		// 生成matcher失败，那么就要把filter从ad 的filters列表中移除
		// 这里 idx+1 在golang里是不做越界检查的，所以不用判断idx+1，真心无语
		ad.Filters = append(ad.Filters[:idx], ad.Filters[idx+1:]...)
	}

}

func AddTagLinkFollowCpl(ad *Ad) {
	// 对于 tag_link button 按钮标签外连广告要增加 TaglinkFollowEnum 匹配条件
	// 用来判断是否关注过，图片所属的用户
	if ad.Type != "button" || ad.Area != "tag_link" {
		return
	}

	tfe := new(xplugin.TaglinkFollowEnum)
	tfe.SetAdId(ad.Id)

	f := &Filter{
		AId:     ad.Id,
		Name:    "taglink_follow_enum",
		Matcher: tfe,
	}

	j, err := simple.NewJson([]byte(ad.Element))
	if err != nil {
		common.Warningf("AddTagLinkFollowCpl adid:%d, adelement:%s failed:%s", ad.Id, ad.Element, err.Error())
		tfe.FollowUid = -1
		ad.Filters = append(ad.Filters, f)
		return
	}

	uid, err := j.Get("uid").Int64()
	if err != nil {
		common.Warningf("AddTagLinkFollowCpl adid:%d, adelement:%s get uid failed:%s", ad.Id, ad.Element, err.Error())
		tfe.FollowUid = -1
		ad.Filters = append(ad.Filters, f)
		return
	}

	tfe.FollowUid = uid
	ad.Filters = append(ad.Filters, f)
}

func IntSlice(str string) ([]int, error) {
	s := strings.Split(str, ".")

	n := make([]int, len(s))
	for i, _ := range s {
		is, err := strconv.Atoi(s[i])
		if err != nil {
			return nil, err
		}
		n[i] = is
	}
	return n, nil
}

func AnalyzValue(data string) (isAll bool, v []string, e error) {
	e = json.Unmarshal(hack.Slice(data), &v)
	if e != nil {
		return
	}

	for i, _ := range v {
		if strings.ToLower(v[i]) == "all" {
			isAll = true
			return
		}
	}
	common.Debugf("AnalyzValue data %s to %v", data, v)
	return
}
