package service

import (
	"golang.org/x/net/context"

	"common"
	"defined"
	"xad"
	"xuser"
)

var _ Command = (*TagDetailAd)(nil)

type TagDetailAd struct {
	BaseAd
}

func init() {
	name := "tag_detail"
	Register(name, func() Command {
		fcd := new(TagDetailAd)
		fcd.SetName(name)
		return fcd
	})
}

func (this *TagDetailAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {

	var mads []*xad.Ad

	mads, err = this.BaseAd.MatchMore(ctx, ui)

	if err == nil {
		for _, ad := range mads {

			ade := GetElement(ctx, ad)
			if ade == nil {
				continue
			}
			if ui.Req.Extra.TagType == nil || ui.Req.Extra.TagID == nil {
				common.Errorf("req not include TagType or TagID")
				continue
			}
			if ade.TagId == *ui.Req.Extra.TagID && ade.TagType == *ui.Req.Extra.TagType {
				ads = append(ads, ad)
			}
		}
	}
	if ads == nil || len(ads) == 0 {
		err = defined.ErrMissingMatchAd
	}
	return
}
