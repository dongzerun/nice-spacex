package service

import (
	"golang.org/x/net/context"

	"common"
	"defined"
	"xad"
	"xuser"
)

var _ Command = (*PasterDetailAd)(nil)

type PasterDetailAd struct {
	BaseAd
}

func init() {
	name := "paster_detail"
	Register(name, func() Command {
		fcd := new(PasterDetailAd)
		fcd.SetName(name)
		return fcd
	})
}

func (this *PasterDetailAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	var mads []*xad.Ad

	mads, err = this.BaseAd.MatchMore(ctx, ui)
	if err != nil {
		return nil, err
	}

	for _, ad := range mads {
		ade := GetElement(ctx, ad)
		if ade == nil {
			continue
		}

		if ui.Req != nil &&
			ui.Req.Extra != nil &&
			ui.Req.Extra.PackageID != nil &&
			ade.PackageId == *ui.Req.Extra.PackageID {
			ads = append(ads, ad)
		} else {
			common.Errorf("missing match package id")
			continue
		}
	}
	if ads == nil || len(ads) == 0 {
		err = defined.ErrMissingMatchAd
	}
	return
}
