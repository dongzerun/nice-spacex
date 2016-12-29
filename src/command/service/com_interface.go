package service

import (
	"golang.org/x/net/context"

	"xad"
	"xuser"
)

var (
	CommandSerivce map[string]Command
)

func init() {
	if CommandSerivce == nil {
		CommandSerivce = make(map[string]Command)
	}
}

// 不同广告位的接口定义
type Command interface {
	Match(context.Context, *xuser.UserInfo) ([]*xad.Ad, error)

	GetName() string
	SetName(string)

	// 获取用户上一次匹配到的广告信息，扔到redis里面
	GetAdByUid(context.Context, int64) (*xad.Ad, error)
	// 保存用户上一次匹配到的广告信息
	// 传入 int64位的uid,和 adid
	SaveAdUid(context.Context, int64, int) error
}

// ad element的通用struct文件
//"{\"package_id\":\"23\",\"package_name\":\"\\u8fd9\\u662f\\u4e2a\\u6d4b\\u8bd5\",\"sid\":\"108\",\"display_position\":\"4\"}"
type AdElement struct {
	Sid       string `json:"sid"`
	Uid       string `json:"uid"`
	PasterId  string `json:"paster_id"`
	PackageId string `json:"package_id"`

	TagId   string `json:"tag_id"`
	TagType string `json:"tag_type"`
}

func Register(name string, fn func() Command) {
	if CommandSerivce == nil {
		CommandSerivce = make(map[string]Command)
	}

	if _, ok := CommandSerivce[name]; ok {
		panic("repeat register command func " + name)
	}
	CommandSerivce[name] = fn()
}
