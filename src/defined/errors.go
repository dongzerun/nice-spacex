package defined

import (
	"errors"
)

var (
	// 获取用户信息的版本号不可用
	ErrVerUnavailable = errors.New("info version unavailable")
	// 未匹配到
	ErrMatcherNotFound = errors.New("matcher not found")
	// ad id 不存在
	ErrAdIdNonExists = errors.New("ad id not exists")
	// 无法根据uid获取上一次的ad信息
	ErrGetAdByUidIellge = errors.New("get ad iellge")
	// 没有匹配到任何广告
	ErrMissingMatchAd = errors.New("missing match ad")

	// matcher 类型断言错误
	ErrMatcherAssert = errors.New("xplugin matcher assert failed")
	// GetMatcher not exists 不存在错误
	ErrMatcherNotExists = errors.New("xplugin matcher not exists")
	// updateFilter AnalyzValue err
	ErrAnalyzValue = errors.New("updateFilter AnalyzValue err")

	// is_new 新注册用户时间戳，必须只有一个值
	ErrIsNewNum = errors.New("updateFilter is_new have not only 1 value")
	// 未实现的plugin
	ErrUnImplementedPlugin = errors.New("plugin un implemented")

	// Rank 排序规则不存在
	ErrRankStrategyNotExists = errors.New("rank stragety not exists")
	ErrAdsAreaNotExists      = errors.New("ads area not exists")

	// context 超时
	ErrContextTimeout = errors.New("context time out")

	ErrFieldsNumUnavailable = errors.New("fields number unavailable")
)
