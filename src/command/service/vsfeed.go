package service

import (
	"golang.org/x/net/context"

	"common"
	"xad"
	"xuser"
)

var _ Command = (*VsFeedCardAd)(nil)

func init() {
	name := "vsfeed_card_3"
	Register(name, func() Command {
		vfc := new(VsFeedCardAd)
		vfc.SetName(name)
		return vfc
	})
}

// 竖滑 Feed 流广告
// 比以前的横滑多一个 运营广告类型，其它正常
type VsFeedCardAd struct {
	BaseAd
}

func (vfc *VsFeedCardAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	ads, err = vfc.BaseAd.Match(ctx, ui)

	if err == nil {
		if len(ads) == 0 {
			return
		}

		// 默认只有会匹配中一个广告，如果是photo类型的需要特殊处理
		ad := ads[0]
		if ad.Type == "photo" {
			// 不符合go的返回值风格
			adE := GetElement(ctx, ad)
			if adE == nil {
				common.Errorf("vsfeed_card_3 photo rematch photo sid undecode failure")
				return
			}

			// 获取tag_link的广告
			tads := getTagLinkAdsBySid(ctx, adE.Sid, ui)
			if tads != nil {
				ads = append(ads, tads)
			}
		}
	}
	return
}
