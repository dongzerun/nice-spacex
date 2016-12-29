package xplugin

import (
	"fmt"

	"golang.org/x/net/context"

	"common"
	"defined"
	"time"
	"xuser"
	"xutil/cache"
)

var _ Matcher = (*DsplyNumEnum)(nil)

// 展示次数匹配
type DsplyNumEnum struct {
	BaseMatch
	Num int // 展现次数
}

func (dne *DsplyNumEnum) IsMatch(ctx context.Context, ui *xuser.UserInfo, rollback chan *RollBack) bool {
	if dne.IsAll() {
		return true
	}

	select {
	case <-ctx.Done():
		common.Warningf("logid:%v DsplyNumEnum timeout or canceled, directly return", ctx.Value("logid"))
		return false
	default:
	}
	var cnt int64
	if ui.Uid == 0 {
		cnt = GetCntByAd(dne.GetAdId(), dne.GetAdStatus(), ui.Req.DeviceID)
	} else {
		common.Debugf("before run get ad:%v user cnt has spent:%v", time.Since(ctx.Value("start_time").(time.Time)).Nanoseconds()/1000)
		cnt = GetCntByAd(dne.GetAdId(), dne.GetAdStatus(), ui.Uid)
	}

	// cnt := GetCntByAdV2(dne.GetAdId(), ui.Uid)
	// common.Infof("GetCntByAdV2 debug ad:%d uid:%d cnt:%d", dne.GetAdId(), ui.Uid, cnt)

	// 注册 rollback 回滚回调函数，对计数做减一操作
	// 改成实实，不需要回滚操作了，由客户端日志来确定计数
	if rollback != nil {
		rollback <- dne.BuildRollBack(ctx, ui)
	}

	return int64(dne.Num) >= cnt
}

// 默认接口不生成回滚函数 return nil
func (dne *DsplyNumEnum) BuildRollBack(ctx context.Context, ui *xuser.UserInfo) *RollBack {
	var (
		key       string
		requestId string
	)

	if ui.Uid == 0 {
		key = fmt.Sprintf(defined.KeyAdUserCntFormat, dne.GetAdId(), dne.GetAdStatus(), ui.Req.DeviceID)
	} else {
		key = fmt.Sprintf(defined.KeyAdUserCntFormat, dne.GetAdId(), dne.GetAdStatus(), ui.Uid)
	}

	// 已经 ctx.Value("logid") 是字符串，所以去掉是否成功断言
	requestId, _ = ctx.Value("logid").(string)

	rb := &RollBack{
		Ad:        dne.GetAdId(),
		RbName:    dne.GetName(),
		Uid:       ui.Uid,
		RequestId: requestId,
	}
	rb.Fn = func() {
		value, err := cache.GlobalRedisCache.IncrBy(key, -1)
		if err == nil && value < 0 {
			// 减为负数了，不应该出现
			cache.GlobalRedisCache.Del(key)
		}
	}
	return rb
}

func GetCntByAd(aid int, status string, unique interface{}) int64 {
	key := fmt.Sprintf(defined.KeyAdUserCntFormat, aid, status, unique)

	value, err := cache.GlobalRedisCache.Incr(key)
	if value == 1 {
		// 说明是第一次计数，需要设置 ttl，自动过期
		err = cache.GlobalRedisCache.SetTTL(key, 86400)
		if err != nil {
			common.Warningf("GetCntByAd setttl failed %s", err.Error())
			return 0
		}
	}
	return value
}

func GetCntByAdV2(aid int, unique interface{}) int64 {
	key := fmt.Sprintf(defined.KeyAdUserCntV2Format, aid, unique)

	value, err := cache.GlobalRedisCache.GetInt64(key)
	if err != nil {
		common.Warningf("GetCntByAdV2 GetInt64 failed %s", err.Error())
	}

	return value
}

func IncrCntByAdV2(aid int, unique interface{}) {
	key := fmt.Sprintf(defined.KeyAdUserCntV2Format, aid, unique)

	value, err := cache.GlobalRedisCache.Incr(key)
	if value == 1 {
		// 说明是第一次计数，需要设置 ttl，自动过期
		err = cache.GlobalRedisCache.SetTTL(key, 86400)
		if err != nil {
			common.Warningf("SetCntByAdV2 setttl failed ", err.Error())
		}
	}
}

// "valid_time", "max_diplay_num", "ios_app_version",
// "android_app_version", "app_channel","gender","city", "tag_list"
func init() {
	name := "max_display_num_enum"
	RegisterMatcher(name, func() Matcher {
		dne := &DsplyNumEnum{}
		dne.BaseMatch.PluginName = name
		return dne
	})
}
