package xuser

import (
	"golang.org/x/net/context"

	"common"
	"defined"
)

type BaseGetter struct {
	Handler
}

func (bg *BaseGetter) UpdateUserInfo(ctx context.Context, ui *UserInfo) error {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v UpdateUserInfo timeout or canceled, directly return", ctx.Value("logid"))
		return defined.ErrContextTimeout
	default:
	}

	var (
		data []byte
		err  error
		key  = bg.Handler.Key(ui)
	)
	data, err = bg.Handler.ReadFromCache(ctx, key)
	if err == nil {
		err = bg.Handler.Decode(ctx, data, ui)
		if err == nil {
			// cache 中查找到数据update ui成功，直接返回
			return nil
		}
		common.Warningf("logid:%v UpdateUserInfo Decode error:%s", ctx.Value("logid"), err.Error())
	}

	common.Warningf("logid:%v UpdateUserInfo ReadFromCache error:%s", ctx.Value("logid"), err.Error())

	//  缓存中数据无效，从落地存储中查找
	err = bg.Handler.ReadFromStorage(ctx, key, ui)
	if err == nil {
		data, err = bg.Handler.Encode(ctx, ui)
		if err == nil {
			bg.Handler.FillCache(ctx, data, ui)
			return nil
		}
		common.Warningf("logid:%v UpdateUserInfo ReadFromStorage FillCache error:%s", ctx.Value("logid"), err.Error())
		return nil
	}
	common.Warningf("logid:%v UpdateUserInfo ReadFromStorage error:%s", ctx.Value("logid"), err.Error())
	return err
}

type DefaultGetter struct {
	BaseGetter
}

type SnappyGetter struct {
	BaseGetter
}

type Lz4Getter struct {
	BaseGetter
}

func NewUiGetter(typ string) UiGetter {

	switch typ {
	case "default":
		ug := new(DefaultGetter)
		ug.BaseGetter.Handler = new(DefaultHandler)
		return ug
	case "lz4":
		ug := new(Lz4Getter)
		ug.BaseGetter.Handler = new(Lz4Handler)
		return ug
	case "snappy":
		ug := new(SnappyGetter)
		ug.BaseGetter.Handler = new(SnappyHandler)
		return ug
	}
	panic("NewUiGetter unknow type " + typ)
}
