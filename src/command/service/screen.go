package service

var _ Command = (*ScreenAd)(nil)

type ScreenAd struct {
	BaseAd
}

func init() {
	name := "enter_screen"
	Register(name, func() Command {
		sad := new(ScreenAd)
		sad.SetName(name)
		return sad
	})
}
