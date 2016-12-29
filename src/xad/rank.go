package xad

import (
	"sort"

	"common"
	"defined"
)

func init() {
	// RegisterRankServer(defined.RankBase, &RankBase{Name: defined.RankBase})
	RegisterRankServer(defined.RankOrder, &RankOrder{Name: defined.RankOrder})
	RegisterRankServer(defined.RankRandom, &RankRandom{Name: defined.RankRandom})
}

// 基类策略, 不排序
type RankBase struct {
	Name string
}

func (rb *RankBase) Rank(ads []*Ad) []*Ad {
	return ads
}

// 基类策略, 按照广告序列排序
// 规则: 线上优先于预上线展示, 广告优先于运营卡片展示
type RankOrder struct {
	Name string
}

func (ro *RankOrder) Rank(ads []*Ad) []*Ad {
	return RankAdsOrder(ads)
}

func RankAdsOrder(ads []*Ad) []*Ad {
	// 全部复制一份
	tmp := make([]*Ad, len(ads))
	for i, _ := range ads {
		tmp[i] = ads[i]
	}
	sort.Sort(AdsOrder(tmp))

	// add debug 打印排序后的广告
	for i, _ := range tmp {
		common.Infof("RankAdsOrder idx:%d ad:%d name:%s area:%s ", i, tmp[i].Id, tmp[i].Name, tmp[i].Area)
	}
	return tmp
}

// 广告排序规则 实现 Len, Less, Swap 三个接口
// 重点在 Less
type AdsOrder []*Ad

func (ao AdsOrder) Len() int {
	return len(ao)
}

// 这么做排序可能不太好，但是暂时够用了
func (ao AdsOrder) Less(i, j int) bool {
	// 排序规则: test广告排在最下面
	if ao[i].Status == "test" && ao[j].Status != "test" {
		return false
	} else if ao[i].Status != "test" && ao[j].Status == "test" {
		return true
	}

	// 排序规则: 运营广告等级低于普通广告，优先展示普通广告
	// 只有 vsfeed_card_3 竖滑才有运营广告，其它场景暂时没有
	if ao[i].OpOrAd == "op" && ao[j].OpOrAd == "ad" {
		return false
	} else if ao[i].OpOrAd == "ad" && ao[j].OpOrAd == "op" {
		return true
	}

	// 排序规则: 最终同样的以随机排列这样比较好
	// 之前是时间排列，现在看不太好，会导致同样广告的展现机会不均等
	// return a[i].UpdateTime < a[j].UpdateTime
	// 随机排序就好
	// fix go1.6 sort.Sort bug
	// if rand.Intn(2) == 0 {
	// 	return true
	// } else {
	// 	return false
	// }

	if ao[i].RankWeight >= ao[j].RankWeight {
		return true
	}

	return false
}

func (ao AdsOrder) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}

// 基类策略, 按照广告序列排序
// 规则: 完全随机,这个策略是给白名单用户看的,完全随机展示广告
type RankRandom struct {
	Name string
}

func (rr *RankRandom) Rank(ads []*Ad) []*Ad {
	return RankAdsRandom(ads)
}

func RankAdsRandom(ads []*Ad) []*Ad {
	tmp := make([]*Ad, len(ads))
	for i, _ := range ads {
		tmp[i] = ads[i]
	}
	sort.Sort(AdsRandom(tmp))

	// add debug 打印排序后的广告
	for i, _ := range tmp {
		common.Infof("RankAdsRandom idx:%d ad:%d name:%s area:%s ", i, tmp[i].Id, tmp[i].Name, tmp[i].Area)
	}
	return tmp
}

type AdsRandom []*Ad

func (ar AdsRandom) Less(i, j int) bool {
	if ar[i].RankWeight >= ar[j].RankWeight {
		return true
	}

	return false
}

func (ar AdsRandom) Len() int {
	return len(ar)
}

func (ar AdsRandom) Swap(i, j int) {
	ar[i], ar[j] = ar[j], ar[i]
}
