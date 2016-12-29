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
)

var (
	_ Command = (*BaseAd)(nil)
)

type BaseAd struct {
	name string
}

func (this *BaseAd) SetName(name string) {
	this.name = name
}

func (this *BaseAd) GetName() string {
	return this.name
}

// 在base层做其实不建义，比较ugly
// 通过 this.GetName() 来判断，有些不妥
// 最好每个实现都有自已的, 实现为空即可，或是直接panic掉？？？
func (this *BaseAd) GetAdByUid(ctx context.Context, uid int64) (*xad.Ad, error) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v GetAdByUid timeout or canceled, directly return", ctx.Value("logid"))
		return nil, defined.ErrContextTimeout
	default:
	}

	var (
		key string
		err error
		ad  *xad.Ad
		id  int
	)

	// 当前只有 vsfeed_card_3 竖滑Feed流才需要跟踪
	if this.GetName() != "vsfeed_card_3" {
		return nil, defined.ErrGetAdByUidIellge
	}

	key = fmt.Sprintf(defined.KeyUidLastAdFormat, uid, this.GetName())
	id, err = cache.GlobalRedisCache.GetInt(key)
	if err != nil {
		return nil, err
	}

	ad, err = xad.AdM.GetAd(id)
	if err != nil {
		common.Warningf("logid:%v GetAdByUid failed requid:%d adid:%d", ctx.Value("logid"), uid, id)
		return nil, err
	}

	return ad, nil
}

func (this *BaseAd) SaveAdUid(ctx context.Context, uid int64, ad int) error {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v SaveAdUid timeout or canceled, directly return", ctx.Value("logid"))
		return defined.ErrContextTimeout
	default:
	}
	// 当前只有 vsfeed_card_3 竖滑Feed流才需要跟踪
	if this.GetName() != "vsfeed_card_3" {
		return nil
	}
	key := fmt.Sprintf(defined.KeyUidLastAdFormat, uid, this.GetName())
	common.Infof("logid:%v SaveAdUid uid:%d key:%s", ctx.Value("logid"), uid, key)
	return cache.GlobalRedisCache.SetWithTTL(key, int64(ad), 86400)
}

// 匹配单个广告
func (this *BaseAd) Match(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	return this.matchMany(ctx, ui, false)
}

//匹配多个广告
func (this *BaseAd) MatchMore(ctx context.Context, ui *xuser.UserInfo) (ads []*xad.Ad, err error) {
	return this.matchMany(ctx, ui, true)
}

func (this *BaseAd) matchMany(ctx context.Context, ui *xuser.UserInfo, isMore bool) ([]*xad.Ad, error) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v matchMany timeout or canceled, directly return", ctx.Value("logid"))
		return nil, defined.ErrContextTimeout
	default:
	}

	common.Infof("logid:%v match name:%s", ctx.Value("logid"), this.GetName())

	ads, err := xad.AdM.MatchAds(ctx, this.GetName(), ui, isMore)
	if err != nil {
		common.Warningf("logid:%v name:%s matchMany err:%s", ctx.Value("logid"), this.GetName(), err.Error())
	}
	return ads, err
}

// decode ad_element 数据
func GetElement(ctx context.Context, ad *xad.Ad) (ade *AdElement) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v GetElement timeout or canceled, directly return", ctx.Value("logid"))
		return nil
	default:
	}

	if ad == nil {
		return nil
	}
	if ad.Element == "" {
		return nil
	}

	var nade AdElement
	common.Debugf("logid:%v ad element:%s", ctx.Value("logid"), ad.Element)
	if err := json.Unmarshal([]byte(ad.Element), &nade); err != nil {
		common.Errorf("logid:%v paster_detail paster id decode err:%s", ctx.Value("logid"), err.Error())
		return nil
	}
	common.Debugf("logid:%v ad element content:%v", ctx.Value("logid"), nade)
	return &nade
}
