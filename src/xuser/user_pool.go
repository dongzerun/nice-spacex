package xuser

import (
	"time"

	"common"
	"config"
	"xutil/pool"
)

var (
	// user getter pool
	userPool *pool.ObjectPool
	// user info struct pool
	userInfoPool *pool.ObjectPool
)

func init() {

	userPool = pool.NewBufferPoolWithSize(10240,
		func() interface{} {
			u := new(User)
			return u
		})

	userInfoPool = pool.NewBufferPoolWithSize(10240,
		func() interface{} {
			ui := &UserInfo{
				TagClass: make([]TagClass, 0, 10),
			}

			return ui
		})

	// 后台不断打印对象池的使用信息
	go PoolStatsDump()
}

func PoolStatsDump() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if userPool != nil {
				common.Infof("Object Pool User stats:%s", userPool.Json())
			}

			if userPool != nil {
				common.Infof("Object Pool UserInfo stats:%s", userPool.Json())
			}
		}
	}
	common.Warningf("PoolStatsDump quit goroutine")
}

func PickUserGetter() *User {
	u := userPool.Get().(*User)
	u.ui = nil
	return u
}

func PutUserGetter(u *User) {
	userPool.Put(u)
}

func PickUserInfo() *UserInfo {
	ui := userInfoPool.Get().(*UserInfo)
	ui.Uid = 0
	ui.Name = ""
	ui.Gender = "secret"
	//ui.Age = 0
	//ui.PlatForm = ""
	//ui.DownloadChannel = ""
	ui.CreateTime = 0
	ui.TagClass = ui.TagClass[:0]
	ui.Req = nil

	if ui.UiGetter == nil {
		ui.UiGetter = NewUiGetter(config.GlobalConfig.MiscConfig.UserCompressed)
	}

	return ui
}

func PutUserInfo(u *UserInfo) {
	userInfoPool.Put(u)
}
