package service

import (
	"golang.org/x/net/context"

	"defined"
	"xad"
	"xuser"
)

var _ Command = (*PasterChosenAd)(nil)

type PasterChosenAd struct {
	BaseAd
}

func init() {
	name := "paster_chosen"
	Register(name, func() Command {
		fcd := new(PasterChosenAd)
		fcd.SetName(name)
		return fcd
	})
}

func (this *PasterChosenAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	var mads []*xad.Ad

	mads, err = this.BaseAd.MatchMore(ctx, ui)

	if err == nil {
		for _, ad := range mads {

			ade := GetElement(ctx, ad)
			if ade == nil {
				continue
			}

			// 把匹配到的全部扔给服务端
			ads = append(ads, ad)
		}
		if ads == nil || len(ads) == 0 { // 完全没有匹配上的值
			err = defined.ErrMissingMatchAd
		}
	}
	return
}
