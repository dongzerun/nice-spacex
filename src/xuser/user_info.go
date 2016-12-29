package xuser

import (
	"math"

	"golang.org/x/net/context"

	"common"
	"defined"
	tf "spacex_tf/spacex"
)

var (
	// 用户数据版本号信息，定期reload
	UserVersion int

	// 最大版本号
	MaxVersion []int

	// 用户白名单
	AdWhiteList map[int64]bool
)

func init() {
	UserVersion = -1

	if MaxVersion == nil {
		MaxVersion := make([]int, 10)
		for i, _ := range MaxVersion {
			MaxVersion[i] = math.MaxInt32
		}
	}

	defined.RegisterOnRun("userinfo", func() {
		common.Info("Server On Run xuser.LoopVersion")
		GetInfoVersion()
		go LoopVersion()
	})
}

type TagClass struct {
	ClassId string `json:"class_id"`
	ClassCn string `json:"class_cn"`
}

type LocationInfo struct {
	Provience  string `json:"provience"`
	Country    string `json:"country"`
	LocationId string `json:"location_id"`
	District   string `json:"district"`
	City       string `json:"city"`
}

// {"name": "-Katze", "gender": "female", "age": "0", "platform": "sina", "download_channel": "", "ctime": "1387532522"}
// ffjson src/xuser/user_info.go
// 这个结构的 json 序列化由 ffjson 手工生成，然后 git add 到工程中
// 结构内容如有修改，请一定要重新生成   user_info_ffjson.go 文件
type UserInfo struct {
	Uid          int64             `json:"uid"`    // 用户ID
	Name         string            `json:"name"`   // 用户名字
	Gender       string            `json:"gender"` // 性别 mail femail secret
	LocationInfo LocationInfo      `json:"location_info"`
	CreateTime   int               `json:"ctime"`     // 创建时间
	TagClass     []TagClass        `json:"tag_class"` // 标签分类
	Req          *tf.RequestParams `json:"-"`         // 请求，带上人的信息
	UiGetter     `json:"-"`        // 更新 userinfo 的接口
}

// 具体实现
type User struct {
	ui *UserInfo
}

func (u *User) PutUserInfo(ui *UserInfo) {
	PutUserInfo(ui)
}

// 获取用户信息的实现，可能在缓存里，不存在才落到 MySQL
func (u *User) GetUserInfo(ctx context.Context, uid int64) (*UserInfo, error) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v GetUserInfo timeout or canceled, directly return", ctx.Value("logid"))
		return nil, defined.ErrContextTimeout
	default:
	}

	ui := PickUserInfo()
	// 填充 uid
	ui.Uid = uid
	// uid == 0，那么此时应该是enter_screen 开屏和 live_h5广告，直接返回就好
	if uid == 0 {
		return ui, nil
	}

	err := ui.UpdateUserInfo(ctx, ui)
	if err != nil {
		common.Warningf("logid:%v GetUserInfo UpdateUserInfo error: %s", ctx.Value("logid"), err.Error())
	}
	u.ui = ui
	common.Infof("logid:%v userinfo:%v", ctx.Value("logid"), ui)
	return ui, nil
}
