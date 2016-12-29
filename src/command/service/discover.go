package service

import (
	simple "github.com/bitly/go-simplejson"
	"golang.org/x/net/context"

	"common"
	"defined"
	"xad"
	"xuser"
	"xutil/hack"
	"xutil/misc"
)

var _ Command = (*DiscoverAd)(nil)

var (
	posSlice = [][]int{[]int{2, 3}, []int{6, 7}, []int{10, 11}}
)

type DiscoverAd struct {
	BaseAd
}

// 分类频道广告元素
type ChannelElement struct {
	AdPhotoUrl  string `json:"ad_photo_url"`
	ChannelName string `json:"channel_name"`
	DisplayPos  int    `json:"display_position"`
	Link        string `json:"link"`
	index       int    `json:"-"`
}

// 推荐广告元素
type RecommendElement struct {
	Sid        string `json:"sid"`
	Link       string `json:"link"`
	DisplayPos int    `json:"display_position"`
}

func (dad *DiscoverAd) Match(ctx context.Context, ui *xuser.UserInfo) ([]*xad.Ad, error) {
	var (
		// 最终广告结果集
		res = make([]*xad.Ad, 0, 7)
		// 分类频道中间集
		ch = make([]*xad.Ad, 0, 6)
		// 为你推荐轮播中间集
		slides = make([]*xad.Ad, 0, 12)
	)
	ads, err := dad.MatchMore(ctx, ui)
	if err != nil {
		return nil, err
	}

	for _, ad := range ads {
		switch ad.Type {
		// channel 1 只存在第一个位置，采用轮播的方式
		case "channel":
			ch = append(ch, ad)
		// 识别广告位轮播 1,2随机选一个，6,8随机选一个
		case "recommend_slides", "recommend_fixed":
			slides = append(slides, ad)
		}
	}

	// 分类频道只取一个广告
	if len(ch) >= 1 {
		res = append(res, ch[0])
	}

	if len(slides) >= 1 {
		res = append(res, pickSlidesAds(ctx, slides, ui)...)
	}

	if len(res) == 0 {
		return nil, defined.ErrMissingMatchAd
	}
	return res, nil
}

// 识别广告位轮播 1,2随机选一个，6,8随机选一个，至多有两个广告
// 20160824 分类频道下线了。在为你推荐这个改为 3或4，7或8，11或12  一共3个广告位。
func pickSlidesAds(ctx context.Context, ads []*xad.Ad, ui *xuser.UserInfo) []*xad.Ad {
	if len(ads) == 0 {
		return nil
	}

	// 待返回的广告集, 至多两个广告
	pickAds := make([]*xad.Ad, 0, 3)
	picked := make([]bool, 3)

	for _, ad := range ads {
		common.Warningf("logid:%v pickSlidesAds start pick ad:%d", ctx.Value("logid"), ad.Id)
		// 已找到两个广告，其它的返回，并回滚计数
		if len(pickAds) == 3 {
			ad.RollBack(ctx, ui)
			common.Warningf("logid:%v pickSlidesAds RollBack ad:%d", ctx.Value("logid"), ad.Id)
			continue
		}
		sj, err := simple.NewJson(hack.Slice(ad.Element))
		if err != nil {
			common.Warningf("logid:%v ad:%d pickSlidesAds NewJson failed:%s ad:%d", ctx.Value("logid"), ad.Id, err.Error(), ad.Id)
			ad.RollBack(ctx, ui)
			continue
		}

		// pos == 1 || pos == 2 || pos == 3
		pos, err := sj.Get("display_position").Int()
		if err != nil {
			common.Warningf("logid:%v ad:%d get display_position failed %s", ctx.Value("logid"), ad.Id, err.Error())
			ad.RollBack(ctx, ui)
			continue
		}

		if pos != 1 && pos != 2 && pos != 3 {
			common.Warningf("logid:%v ad:%d get display_position not 1 2 or 3 %d", ctx.Value("logid"), ad.Id, pos)
			ad.RollBack(ctx, ui)
			continue
		}

		if picked[pos-1] == true {
			common.Warningf("logid:%v ad:%d dup display_position %d, just rollback", ctx.Value("logid"), ad.Id, pos)
			ad.RollBack(ctx, ui)
			continue
		} else {
			picked[pos-1] = true
		}

		sj.Set("display_position", posSlice[pos-1][misc.RandomInt(2)])

		data, err := sj.MarshalJSON()
		if err != nil {
			common.Warningf("logid:%v ad:%d assertPos MarshalJSON failed %s", ctx.Value("logid"), ad.Id, err.Error())
			ad.RollBack(ctx, ui)
			continue
		}

		// 深拷贝
		destAd := xad.DeepCopyAd(ad)
		destAd.Element = hack.String(data)

		pickAds = append(pickAds, destAd)
	}

	return pickAds
}

func assertElementPos(ad *xad.Ad, pos int, sj *simple.Json) {
	if sj == nil {
		var err error
		sj, err = simple.NewJson(hack.Slice(ad.Element))
		if err != nil {
			common.Warningf("assertPos NewJson failed %s", err.Error())
			// 忽略
			return
		}
	}

	curpos, err := sj.Get("display_position").Int()
	if err != nil {
		common.Warningf("assertElementPos simple json get display_position failed adelement:%s", ad.Element)
	}
	if curpos != pos {
		sj.Set("display_position", 1)
	}

	data, err := sj.MarshalJSON()
	if err != nil {
		common.Warningf("assertPos MarshalJSON failed %s", err.Error())
		// 忽略
		return
	}

	ad.Element = hack.String(data)
}

func init() {
	// 发现页广告匹配
	// channel：分类频道，recommend_fixed：为你推荐固定位置，recommend_slides:为你推荐轮播位置
	// 以数组形式取出来，全部返回给服务端
	name := "discover_index"
	Register(name, func() Command {
		dad := new(DiscoverAd)
		dad.SetName(name)
		return dad
	})
}
