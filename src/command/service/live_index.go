package service

import (
	"golang.org/x/net/context"

	// "common"
	"defined"
	"xad"
	"xuser"
)

var _ Command = (*LiveIndexAd)(nil)

type LiveIndexAd struct {
	BaseAd
}

func (lia *LiveIndexAd) Match(ctx context.Context, ui *xuser.UserInfo) ([]*xad.Ad, error) {
	ads, err := lia.MatchMore(ctx, ui)
	if err != nil {
		return nil, err
	}

	if len(ads) == 0 {
		return nil, defined.ErrMissingMatchAd
	}
	return ads, nil
}

func init() {
	// 直播发现页广告
	// ad_area='live_index' ad_type=live_fixed|banner_slides
	name := "live_index"
	Register(name, func() Command {
		lia := new(LiveIndexAd)
		lia.SetName(name)
		return lia
	})
}
