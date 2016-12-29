package xad

import (
	"golang.org/x/net/context"

	"common"
	"defined"
	"xad/xplugin"
	"xuser"
)

var (
	// 所有的area 类型
	AdAreas []string

	// 全局广告manager
	AdM *AdManager
)

func init() {
	if AdAreas == nil {
		AdAreas = make([]string, 0, 3)
	}

	if AdM == nil {
		AdM = NewAdMana()
	}

	defined.RegisterOnRun("xad", func() {
		common.Info("Server On Run xad.InitAd")
		InitAd()
	})
}

func InitAd() {
	AdAreas = GetAllAreas()
	if len(AdAreas) == 0 {
		common.Warning("InitAd first load ad areas error")
	}

	common.Info("InitAd all GetAds")
	GetAds()

	common.Info("InitAd start LoopReloadAds")
	go LoopReloadAds()
}

// 广告结构体，最后一个字段用来区分是普通广告还是运营广告
// 目前只有竖滑 Feed 流才有运营广告，其它过滤策略相同
type Ad struct {
	Id          int       `json:"ad_id"`          // 广告ID
	Name        string    `json:"ad_name"`        // 广告名称
	Description string    `json:"ad_description"` // 广告描述
	Area        string    `json:"ad_area"`        // 广告位置 enum("enter_screen", "feed_card_3", "tag_link")
	Type        string    `json:"ad_type"`        // 广告类型 user photo tag paster link dynamic static tag_link button
	Element     string    `json:"ad_element"`     // 广告元素 json 数据
	Status      string    `json:"status"`         // 广告状态 enum("online", "pause", "offline", "test")
	CreateTime  int       `json:"create_time"`    // 创建时间
	UpdateTime  int       `json:"update_time"`    // 更新时间
	Filters     []*Filter `json:"-"`              // 广告过滤条件集合 and 关系
	OpOrAd      string    `json:"op_or_ad"`       // 标记是运营卡片还是广告卡片
	WhiteList   []int     `json:"-"`              // 白名单里面的人无次数限制
	RankWeight  int       `json:"-"`              // 只是为了做权重排序
}

func DeepCopyAd(src *Ad) *Ad {
	return &Ad{
		Id:          src.Id,
		Name:        src.Name,
		Description: src.Description,
		Area:        src.Area,
		Type:        src.Type,
		Element:     src.Element,
		Status:      src.Status,
		CreateTime:  src.CreateTime,
		UpdateTime:  src.UpdateTime,
		OpOrAd:      src.OpOrAd,
		WhiteList:   src.WhiteList,
		RankWeight:  src.RankWeight,
	}
}

// 按照一定规则编排广告的过滤模块，暂时未实现，预留即可
// 广告初始化时执行一次
func (ad *Ad) arrange() {

}

// 获取广告剩余有效时间
func (ad *Ad) remaining(ctx context.Context) int {

	for _, f := range ad.Filters {
		if f.Name != "valid_time" {
			continue
		}

		v, ok := f.Matcher.(*xplugin.ValidTimeRange)
		if ok {
			return v.Remaining(ctx)
		}

		return 0
	}

	return 0
}

// 具体的匹配模块
func (ad *Ad) match(ctx context.Context, ui *xuser.UserInfo, rollback chan *xplugin.RollBack) bool {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v match  Ad:%d timeout or canceled, directly return", ctx.Value("logid"), ad.Id)
		return false
	default:
	}

	// 对于广告内部白名单用户，屏蔽一切过滤条件，无条件展示
	for _, u := range ad.WhiteList {
		if ui.Uid == int64(u) {
			// 命中白名单用户, 只需检测有效时间，其它全部忽略即可
			common.Infof("logid:%v uid:%d match inner white list", ctx.Value("logid"), ui.Uid)
			for _, filter := range ad.Filters {
				if filter.Name == "valid_time" {
					return filter.IsMatch(ctx, ui, rollback)
				}
			}
		}
	}

	isMatchAll := true
	for _, filter := range ad.Filters {
		if !filter.IsMatch(ctx, ui, rollback) {
			common.Warningf("logid:%v ad:%d name:%s missing match:%s", ctx.Value("logid"), ad.Id, ad.Name, filter.Name)
			isMatchAll = false
			break
		}
	}

	return isMatchAll
}

// 广告回滚资源
func (ad *Ad) RollBack(ctx context.Context, ui *xuser.UserInfo) {
	for _, cpl := range ad.Filters {
		rb := cpl.Matcher.BuildRollBack(ctx, ui)
		select {
		case GlobalCancelChan <- rb:
			common.Warningf("logid:%v RollBack ad:%d xplugin:%s", ctx.Value("logid"), ad.Id, cpl.Name)
		default:
			// 下面的日志理论上一条都不应用，有则说明异步消费 callback慢了
			common.Warningf("logid:%v rollback rejected ad:%d", ctx.Value("logid"), rb.Ad)
		}
	}
}

// 广告回滚计数资源
func (ad *Ad) RollBackDisplayNum(ctx context.Context, ui *xuser.UserInfo) {
	for _, cpl := range ad.Filters {
		dsply, ok := cpl.Matcher.(*xplugin.DsplyNumEnum)
		if ok {
			rb := dsply.BuildRollBack(ctx, ui)
			select {
			case GlobalCancelChan <- rb:
				common.Warningf("logid:%v RollBack ad:%d xplugin:%s", ctx.Value("logid"), ad.Id, cpl.Name)
			default:
				// 下面的日志理论上一条都不应用，有则说明异步消费 callback慢了
				common.Warningf("logid:%v rollback rejected ad:%d", ctx.Value("logid"), rb.Ad)
			}
		}
	}
}
