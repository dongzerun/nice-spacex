package xuser

import (
	"golang.org/x/net/context"
)

type UiGetter interface {
	UpdateUserInfo(context.Context, *UserInfo) error
}

// 解压缩接口
type Handler interface {
	// 压缩
	Encode(context.Context, *UserInfo) ([]byte, error)
	// 解压
	Decode(context.Context, []byte, *UserInfo) error
	// 生成 cache key
	Key(*UserInfo) string
	// 从缓存中读数据
	ReadFromCache(context.Context, string) ([]byte, error)
	// 从存储中读数据并更新 userinfo
	ReadFromStorage(context.Context, string, *UserInfo) error
	// 填充一级和二级cache
	FillCache(context.Context, []byte, *UserInfo)
}
