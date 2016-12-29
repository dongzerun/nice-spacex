package xad

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"defined"
	"xuser"
)

// AdManager 当前实现简单
// 以后后涉及多个广告的流量分配, 更多的匹配规则
type AdManager struct {
	// 一个大map和全局锁暂时够用了, 以后广告再多就需要打散成多个buckets
	// 第一个string是rank策略, 当前有rank_order和rank_random
	// 第二个string是广告位, 比如feed_card_3,discover_index
	ads  map[string]map[string][]*Ad
	lock sync.RWMutex
}

func NewAdMana() *AdManager {
	adm := &AdManager{
		ads: make(map[string]map[string][]*Ad),
	}
	return adm
}

// type是广告位置, rank是排序策略
func (am *AdManager) Get(typ string, rank string) ([]*Ad, error) {
	var (
		ok      bool
		ads     []*Ad
		rankAds map[string][]*Ad
	)
	am.lock.RLock()
	defer am.lock.RUnlock()

	rankAds, ok = am.ads[rank]
	if !ok || rankAds == nil {
		return nil, defined.ErrRankStrategyNotExists
	}

	ads, ok = rankAds[typ]
	if !ok {
		return nil, defined.ErrAdsAreaNotExists
	}
	return ads, nil
}

// 设置的时候，要把所有策略的都设置一遍
func (am *AdManager) Set(typ string, ads []*Ad) {
	am.lock.Lock()
	defer am.lock.Unlock()

	for name, rankserver := range RankServer {
		rankmap, ok := am.ads[name]
		if !ok {
			rankmap = make(map[string][]*Ad)
			am.ads[name] = rankmap
		}

		rankmap[typ] = rankserver.Rank(ads)
	}
}

// 删除一个区域的广告时, 要把所有策略的都删除掉
func (am *AdManager) Del(typ string) {
	am.lock.Lock()
	defer am.lock.Unlock()

	for _, areamap := range am.ads {
		if areamap != nil {
			delete(areamap, typ)
		}
	}
}

func (am *AdManager) GetAd(id int) (*Ad, error) {
	am.lock.RLock()
	defer am.lock.RUnlock()

	// 只须遍历一个策略的，因为每个策略都包括全量广告
	for _, areamap := range am.ads {
		if areamap == nil {
			continue
		}

		for _, ads := range areamap {
			for _, ad := range ads {
				if ad.Id == id {
					return ad, nil
				}
			}
		}

		return nil, defined.ErrAdIdNonExists
	}
	return nil, defined.ErrAdIdNonExists
}

func (am *AdManager) MatchAds(ctx context.Context, area string, ui *xuser.UserInfo, isMore bool) ([]*Ad, error) {
	select {
	case <-ctx.Done():
		return nil, defined.ErrContextTimeout
	default:
	}

	var (
		adverts []*Ad
		err     error
		ads     []*Ad
	)

	// 直播聚合页 live_h5 live_collection_page 忽略所有过滤条件
	if area == "live_h5" {
		return am.Get(area, defined.RankRandom)
	}

	if xuser.IsInPreOnlineList(ui.Uid) {
		adverts, err = am.Get(area, defined.RankRandom)
	} else {
		adverts, err = am.Get(area, defined.RankOrder)
	}

	if err != nil {
		return nil, err
	}

	if len(adverts) == 0 {
		return nil, fmt.Errorf("area:%s ad is null", area)
	}

	exec := PickExecutor()
	ads, err = exec.MultiGo(ctx, ui, adverts, isMore)
	PutExecutor(exec)

	return ads, err
}

// 获取广告有效时间,一把大锁撸遍全场
func (am *AdManager) Remaining(ctx context.Context, id int) (int, error) {
	select {
	case <-ctx.Done():
		return 0, defined.ErrContextTimeout
	default:
	}

	ad, err := am.GetAd(id)
	if err != nil {
		return 0, err
	}
	return ad.remaining(ctx), nil
}
