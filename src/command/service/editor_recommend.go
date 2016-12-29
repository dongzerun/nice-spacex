package service

var _ Command = (*EditorRecommendAd)(nil)

type EditorRecommendAd struct {
	BaseAd
}

func init() {
	name := "editor_recommend"
	Register(name, func() Command {
		fcd := new(EditorRecommendAd)
		fcd.SetName(name)
		return fcd
	})
}
