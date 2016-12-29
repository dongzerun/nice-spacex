package service

var _ Command = (*UserGuideAd)(nil)

func init() {
	name := "user_guide"
	Register(name, func() Command {
		ug := new(UserGuideAd)
		ug.SetName(name)
		return ug
	})
}

// 用户引导类型卡片
type UserGuideAd struct {
	BaseAd
}
