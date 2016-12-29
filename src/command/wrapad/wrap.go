package wrapad

import (
	"xad"

	"golang.org/x/net/context"

	tf "spacex_tf/spacex"
)

// 注册回调函数，对AD进行更改, 第一个参数是ad,第二个是request
var PostFuncs map[string]func(context.Context, *xad.Ad, *tf.RequestParams, string)

func init() {
	if PostFuncs == nil {
		PostFuncs = make(map[string]func(context.Context, *xad.Ad, *tf.RequestParams, string))
	}
}

func RegisterPostFunc(n string, fn func(context.Context, *xad.Ad, *tf.RequestParams, string)) {
	if PostFuncs == nil {
		PostFuncs = make(map[string]func(context.Context, *xad.Ad, *tf.RequestParams, string))
	}
	_, exists := PostFuncs[n]
	if exists {
		panic("PostFunc " + n + " already registed")
	}

	PostFuncs[n] = fn
}
