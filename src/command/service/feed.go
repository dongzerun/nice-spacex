package service

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	"xad"
	"xuser"
	"xutil/cache"
	"xutil/hack"
)

var _ Command = (*FeedCardAd)(nil)

type FeedCardAd struct {
	BaseAd
}

func init() {
	name := "feed_card_3"
	Register(name, func() Command {
		fcd := new(FeedCardAd)
		fcd.SetName(name)
		return fcd
	})
}

func (this *FeedCardAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	ads, err = this.BaseAd.Match(ctx, ui)

	if err == nil {
		if len(ads) == 0 {
			return
		}
		ad := ads[0] // 默认只有会匹配中一个广告
		if ad.Type != "photo" {
			return
		}

		var adE AdElement
		merr := json.Unmarshal([]byte(ad.Element), &adE)
		if merr != nil {
			common.Errorf("feed_card_3 photo rematch photo sid undecode err:%s", merr.Error())
			return
		}

		// 获取tag_link的广告
		common.Infof("logid:%v feedcard photo start get taglink", ctx.Value("logid"))
		tads := getTagLinkAdsBySid(ctx, adE.Sid, ui)
		if tads != nil {
			ads = append(ads, tads)
		}
	}
	return
}

// 获取照片的外链广告
func getTagLinkAdsBySid(ctx context.Context, sid string, ui *xuser.UserInfo) (rad *xad.Ad) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v getTagLinkAdsBySid timeout or canceled, directly return", ctx.Value("logid"))
		return nil
	default:
	}

	matchAds := make([]*xad.Ad, 0)

	// feed card 下面的button && tag_link匹配
	ads, err := xad.AdM.MatchAds(ctx, "tag_link", ui, true)
	if err != nil {
		common.Warningf("logid:%v name:%s matchMany err:%s", ctx.Value("logid"), "getTagLinkAdsBySid", err.Error())
		return
	}

	for _, ad := range ads {
		if ad.Type != "button" && ad.Type != "tag_link" {
			common.Warningf("logid:%v adid:%d not button or tag_link get:%s", ctx.Value("logid"), ad.Id, ad.Type)
			continue
		}

		ade := GetElement(ctx, ad)
		if ade == nil || ade.Sid != sid { // 根据照片id 去拿标签外链 | 标签按钮
			continue
		}
		matchAds = append(matchAds, ad)
	}

	if len(matchAds) > 0 {
		for _, sad := range matchAds { // 遍历广告，寻找是已经命中过的广告id
			key := fmt.Sprintf(defined.KeyTagLinkADCacheFormat, sad.Id, ui.Uid)

			if IsTagLinkIsInCache(key) { // 如果已经命中过该tag_link 广告，则继续返回tag_link广告
				common.Debugf("logid:%v having match id:%d redis key:%v", ctx.Value("logid"), sad.Id, key)
				return sad
			}
		}
		// 只会返回随机的一个广告
		mad := matchAds[0]

		common.Debugf("logid:%v feed card random ad:%v", ctx.Value("logid"), mad.Id)

		if mad.Type == "tag_link" { // 如果是tag link就cache
			common.Debugf("logid:%v feed card random ad type:%v", ctx.Value("logid"), mad.Type)
			key := fmt.Sprintf(defined.KeyTagLinkADCacheFormat, mad.Id, ui.Uid)
			cache.GlobalRedisCache.SetByteWithEx(key, hack.Slice(""), 86400)
		}
		return mad
	}
	common.Warningf("logid:%v get null TagLinkAds", ctx.Value("logid"))

	return
}
