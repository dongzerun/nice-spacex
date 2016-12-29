package xuser

import (
	"encoding/json"
	"fmt"

	lz4 "github.com/bkaradzic/go-lz4"
	"github.com/golang/snappy"
	"golang.org/x/net/context"

	"common"
	"defined"
	"xutil/cache"
	"xutil/hack"
)

// 默认情况下不压缩
type DefaultHandler struct {
}

func (dh *DefaultHandler) Encode(ctx context.Context, ui *UserInfo) ([]byte, error) {
	return json.Marshal(ui)
}

func (dh *DefaultHandler) Decode(ctx context.Context, d []byte, ui *UserInfo) error {

	return json.Unmarshal(d, ui)
}

func (dh *DefaultHandler) Key(ui *UserInfo) string {
	// 默认是未经压缩
	return fmt.Sprintf(defined.KeyUICacheFormat, ui.Uid)
}

func (dh *DefaultHandler) ReadFromCache(ctx context.Context, key string) ([]byte, error) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v ReadFromCache timeout or canceled, directly return", ctx.Value("logid"))
		return nil, defined.ErrContextTimeout
	default:
	}

	var (
		data   []byte
		exists bool
		err    error
	)
	// 第一步尝试从程序内置的 LRU Cache 中找到数据
	data, exists = lru.Get(key)
	if exists {
		return data, nil
	}

	// 第二步尝试从二级 redis cache 中查找
	data, err = cache.GlobalRedisCache.GetByte(key)
	// 冲一级程序内置 cache
	if err == nil {
		lru.Set(key, data)
		return data, nil
	}
	// cache 查找失效，等待存储中查询
	return nil, err
}

// ReadFromStorage 从MySQL中读取信息后，更新 usrinfo
func (dh *DefaultHandler) ReadFromStorage(ctx context.Context, key string, ui *UserInfo) error {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v ReadFromStorage timeout or canceled, directly return", ctx.Value("logid"))
		return defined.ErrContextTimeout
	default:
	}

	im, mu, err := GetUserInfo(ctx, ui.Uid)
	if err != nil {
		common.Warningf("logid:%v ReadFromStorage from DB uid:%d err:%s", ctx.Value("logid"), ui.Uid, err.Error())
		return err
	}

	// common.Debugf("logid:%v ReadFromStorage im:%s mu:%s", ctx.Value("logid"), im, mu)

	err = json.Unmarshal(hack.Slice(im), ui)
	if err != nil {
		common.Warningf("logid:%v ReadFromStorage uid:%d unmarshal im:%s err:%s", ctx.Value("logid"), ui.Uid, im, err.Error())
		return err
	}

	err = json.Unmarshal(hack.Slice(mu), ui)
	if err != nil {
		// 可变的用户信息包括tags，有些大V是有问题的，暂时忽略先匹配上广告再说
		common.Warningf("logid:%v ReadFromStorage uid:%d unmarshal mu:%s err:%s just ignore mutable info", ctx.Value("logid"), ui.Uid, mu, err.Error())
		// return err
	}

	return nil
}

func (dh *DefaultHandler) FillCache(ctx context.Context, data []byte, ui *UserInfo) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v FillCache timeout or canceled, directly return", ctx.Value("logid"))
		return
	default:
	}

	key := dh.Key(ui)
	lru.Set(key, data)

	err := cache.GlobalRedisCache.SetByteWithEx(key, data, 3600*12)
	if err != nil {
		common.Warningf("logid:%v FillCache uid:%d SetByteWithEx err:%s ", ctx.Value("logid"), ui.Uid, err.Error())
	}
	common.Infof("logid:%v FillCache uid:%d success ", ctx.Value("logid"), ui.Uid)
}

// 使用 snappy 来压缩
type SnappyHandler struct {
	DefaultHandler
}

func (sh *SnappyHandler) Key(ui *UserInfo) string {
	// 默认是未经压缩
	return fmt.Sprintf(defined.KeyUICacheSnappyFormat, ui.Uid)
}

func (sh *SnappyHandler) Encode(ctx context.Context, ui *UserInfo) ([]byte, error) {
	data, err := json.Marshal(ui)
	if err != nil {
		return nil, err
	}

	compressed := snappy.Encode(nil, data)
	uncLen := len(data)
	cLen := len(compressed)
	common.Infof("SnappyHandler uncompressed:%d compressed:%d, ratio:%.2f", uncLen, cLen, float64(uncLen)/float64(cLen))
	return compressed, nil
}

func (sh *SnappyHandler) Decode(ctx context.Context, d []byte, ui *UserInfo) error {
	data, err := snappy.Decode(nil, d)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ui)
}

// 使用 lz4 来压缩
type Lz4Handler struct {
	DefaultHandler
}

func (lh *Lz4Handler) Key(ui *UserInfo) string {
	// 默认是未经压缩
	return fmt.Sprintf(defined.KeyUICacheLz4Format, ui.Uid)
}

func (lh *Lz4Handler) Encode(ctx context.Context, ui *UserInfo) ([]byte, error) {
	data, err := json.Marshal(ui)
	if err != nil {
		return nil, err
	}

	compressed, e := lz4.Encode(nil, data)
	if e != nil {
		return nil, e
	}
	uncLen := len(data)
	cLen := len(compressed)
	common.Infof("Lz4Handler uncompressed:%d compressed:%d, ratio:%.2f", uncLen, cLen, float64(uncLen)/float64(cLen))
	return compressed, nil
}

func (lh *Lz4Handler) Decode(ctx context.Context, d []byte, ui *UserInfo) error {
	data, err := lz4.Decode(nil, d)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ui)
}
