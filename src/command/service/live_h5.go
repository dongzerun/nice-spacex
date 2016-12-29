package service

import (
	"golang.org/x/net/context"

	// "common"
	"defined"
	"xad"
	"xuser"
)

var _ Command = (*LiveH5Ad)(nil)

type LiveH5Ad struct {
	BaseAd
}

func (lha *LiveH5Ad) Match(ctx context.Context, ui *xuser.UserInfo) ([]*xad.Ad, error) {
	ads, err := lha.MatchMore(ctx, ui)
	if err != nil {
		return nil, err
	}

	if len(ads) == 0 {
		return nil, defined.ErrMissingMatchAd
	}
	return ads, nil
}

func init() {
	// 直播H5发现页广告
	name := "live_h5"
	Register(name, func() Command {
		lha := new(LiveH5Ad)
		lha.SetName(name)
		return lha
	})
}
