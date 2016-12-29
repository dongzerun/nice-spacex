package feedback

import (
	"golang.org/x/net/context"

	tf "spacex_tf/spacex"
)

// 点击反馈事件处理注册表
var FeedServiceMap map[string]Action

func init() {
	if FeedServiceMap == nil {
		FeedServiceMap = make(map[string]Action)
	}
}

type Action interface {
	GetName() string
	SetName(string)
	Process(context.Context, *tf.FeedbackParams) *tf.Response
}

type Base struct {
	name string
}

func (b *Base) SetName(n string) {
	b.name = n
}

func (b *Base) GetName() string {
	return b.name
}

func Register(action string, fn func() Action) {
	if FeedServiceMap == nil {
		FeedServiceMap = make(map[string]Action)
	}

	if _, ok := FeedServiceMap[action]; ok {
		panic("repeat register feed func " + action)
	}
	FeedServiceMap[action] = fn()
}
