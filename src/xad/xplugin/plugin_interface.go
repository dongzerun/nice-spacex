package xplugin

import (
	"fmt"

	"golang.org/x/net/context"

	"defined"
	"xuser"
)

var (
	MatcherMap map[string]func() Matcher
)

type RollBack struct {
	Ad        int    `json:"id"`            // 广告唯一标识 ID
	RbName    string `json:"rollback_name"` // 回滚模块名称
	Uid       int64  `json:"uid"`           // 当前回滚针对的人
	RequestId string `json:"request_id"`    // 当前请求唯一ID
	Fn        func() `json:"-"`             // 具体回滚方法回调
}

// 广告与用户的匹配模式接口
type Matcher interface {
	// 是否匹配统一暴露给外面最后使用的函数
	IsMatch(context.Context, *xuser.UserInfo, chan *RollBack) bool

	// 是否全部匹配
	IsAll() bool

	// 设置全部匹配
	SetAll()

	// 设置 Flase
	IsFalse() bool
	SetFalse()

	// 返回模块名称
	GetName() string

	// 设置广告ID
	SetAdId(int)
	GetAdId() int

	// 设置文告状态
	SetAdStatus(status string)
	GetAdStatus() string

	// 设置广告类型
	SetAdArea(string)
	GetAdArea() string

	// 设置白名单
	GetWL() []int
	SetWL(wl []int)

	// 生成 rollback 回调函数
	BuildRollBack(context.Context, *xuser.UserInfo) *RollBack
}

type BaseMatch struct {
	PluginName string
	All        bool
	None       bool
	AdId       int
	AdStatus   string // 广告状态
	AdArea     string // 广告位置
	WhiteList  []int  // 白名单匹配，不受display num限制
}

func (bm *BaseMatch) SetWL(wl []int) {
	bm.WhiteList = wl
}

func (bm *BaseMatch) GetWL() []int {
	return bm.WhiteList
}

func (bm *BaseMatch) SetAdStatus(status string) {
	bm.AdStatus = status
}

func (bm *BaseMatch) GetAdStatus() string {
	return bm.AdStatus
}

func (bm *BaseMatch) SetAdId(id int) {
	bm.AdId = id
}

func (bm *BaseMatch) GetAdId() int {
	return bm.AdId
}

func (bm *BaseMatch) SetAdArea(area string) {
	bm.AdArea = area
}

func (bm *BaseMatch) GetAdArea() string {
	return bm.AdArea
}

func (bm *BaseMatch) IsAll() bool {
	return bm.All
}

func (bm *BaseMatch) SetAll() {
	bm.All = true
}

func (bm *BaseMatch) IsFalse() bool {
	return bm.None == true
}

func (bm *BaseMatch) SetFalse() {
	bm.None = true
}

func (bm *BaseMatch) GetName() string {
	return bm.PluginName
}

// 默认接口不生成回滚函数 return nil
func (bm *BaseMatch) BuildRollBack(ctx context.Context, ui *xuser.UserInfo) *RollBack {
	return nil
}

func GetMatcher(name, typ string) (Matcher, error) {
	key := fmt.Sprintf("%s_%s", name, typ)
	fn, ok := MatcherMap[key]
	if !ok {
		return nil, defined.ErrMatcherNotFound
	}
	return fn(), nil
}

func RegisterMatcher(name string, fn func() Matcher) {
	if MatcherMap == nil {
		MatcherMap = make(map[string]func() Matcher)
	}
	if _, ok := MatcherMap[name]; ok {
		panic("matcher already exists " + name)
	}
	MatcherMap[name] = fn
}
