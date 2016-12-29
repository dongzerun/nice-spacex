package service

import (
	"encoding/json"
	"strconv"

	"golang.org/x/net/context"

	"common"
	"defined"
	"xad"
	"xad/xplugin"
	"xuser"
)

var _ Command = (*NoticeAd)(nil)

type NoticeAd struct {
	BaseAd
}

type NoticeElement struct {
	Uid string
}

func (na *NoticeAd) Match(ctx context.Context, ui *xuser.UserInfo) ([]*xad.Ad, error) {
	var (
		mads []*xad.Ad
		ads  = make([]*xad.Ad, 0, 1)
		err  error
		uid  int
	)

	mads, err = na.BaseAd.MatchMore(ctx, ui)
	if err != nil {
		return nil, err
	}

	for _, ad := range mads {
		var e NoticeElement
		err := json.Unmarshal([]byte(ad.Element), &e)
		if err != nil {
			common.Warningf("logid:%s notice unmarshal element:%s error:%s", ctx.Value("logid"), ad.Element, err.Error())
			continue
		}

		uid, err = strconv.Atoi(e.Uid)
		if err != nil {
			common.Warningf("logid:%s notice atoi uid:%s error:%s", ctx.Value("logid"), e.Uid, err.Error())
			continue
		}
		// 品牌自已不推送
		if ui.Uid == int64(uid) {
			continue
		}
		// todo:
		// ugly 代码侵入太强 下一个版本一定干掉这块 做好MVC分层
		// 只推送未关注的
		if xplugin.IsTagLinkFollow(ctx, ui.Uid, int64(uid)) {
			continue
		}

		// 把匹配到的全部扔给服务端
		ads = append(ads, ad)
	}

	// 完全没有匹配上的值
	if ads == nil || len(ads) == 0 {
		return nil, defined.ErrMissingMatchAd
	}
	return ads, nil
}

func init() {
	name := "notice"
	Register(name, func() Command {
		nad := new(NoticeAd)
		nad.SetName(name)
		return nad
	})
}
