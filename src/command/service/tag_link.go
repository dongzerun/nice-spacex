package service

import (
	// "fmt"
	"golang.org/x/net/context"

	"defined"
	"xad"
	"xuser"
	"xutil/cache"
)

var _ Command = (*TagLinkAd)(nil)

type TagLinkAd struct {
	BaseAd
}

func init() {
	name := "tag_link"
	Register(name, func() Command {
		tla := new(TagLinkAd)
		tla.SetName(name)
		return tla
	})
}

// 匹配单个广告
func (this *TagLinkAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	ads, err = this.matchMany(ctx, ui, true)
	if len(ads) == 0 {
		err = defined.ErrMissingMatchAd
		return
	}

	// for _, ad := range mads {
	// 	key := fmt.Sprintf(defined.KeyTagLinkADCacheFormat, ad.Id, ui.Uid)
	// 	if IsTagLinkIsInCache(key) && ad.Type == "tag_link" { // 对于已经命中tag_link广告缓存，并且为tag_link
	// 		ads = append(ads, ad)
	// 		return
	// 	}
	// }
	// err = defined.ErrMissingMatchAd
	return
}

func IsTagLinkIsInCache(key string) bool {
	if is, _ := cache.GlobalRedisCache.Exists(key); is {
		return true
	}
	return false
}
