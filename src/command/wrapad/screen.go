package wrapad

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/context"

	"common"
	tf "spacex_tf/spacex"
	"xad"
)

func init() {
	RegisterPostFunc("screen_modified", ScreenModifiedByDeviceType)
}

// {"ad_photo_url":"http:\/\/img08.oneniceapp.com\/upload\/share\/e29d6528849a24f171c84488a46ca86d.gif",
// "link":"","stay":"5","min_interval":"8","start":1470197945,
// "finish":1470284345,"max_display_num":"10"}
// "photo_info":[{"size":1,"url":"XX"},{"size":2,"url":"XX"},{"size":3,"url":"XXX"}]
type ScreenElement struct {
	AdPhotoUrl    string       `json:"ad_photo_url"`
	JointPhotoUrl string       `json:"joint_photo_url"`
	Link          string       `json:"link"`
	Stay          string       `json:"stay"`
	BrandStay     string       `json:"brand_stay"`
	MinInterval   string       `json:"min_interval"`
	Start         int          `json:"start"`
	Finish        int          `json:"finish"`
	MaxDisplayNum string       `json:"max_display_num"`
	PhotoInfos    []*PhotoInfo `json:"photo_info"`
}

type PhotoInfo struct {
	Size     int    `json:"size"`
	Url      string `json:"url"`       //品牌广告url
	JointUrl string `json:"joint_url"` //品牌联名广告url
}

// 只对开屏有效，匹配手机机型，生成对应尺寸的图片url
func ScreenModifiedByDeviceType(ctx context.Context, ad *xad.Ad, req *tf.RequestParams, source string) {
	// 该wrap仅对开屏有效，所以直接退出
	if ad.Area != "enter_screen" {
		return
	}

	var (
		se       *ScreenElement
		err      error
		size     int
		setPhoto bool
		data     []byte
	)

	se = &ScreenElement{
		PhotoInfos: make([]*PhotoInfo, 0, 3),
	}

	err = json.Unmarshal([]byte(ad.Element), se)
	if err != nil {
		common.Warningf("logid:%v ScreenModified unmarshal %s failed error %s", ctx.Value("logid"), ad.Element, err.Error())
		return
	}

	size = getPhotoType(req)
	if size < 1 || size > 3 {
		common.Warningf("logid:%v ScreenModified getPhotoType get %d unavailable,use 2", ctx.Value("logid"), size)
		size = 2
	}

	// 设置匹配的url
	for _, pi := range se.PhotoInfos {
		if pi.Size == size {
			se.AdPhotoUrl = pi.Url
			se.JointPhotoUrl = pi.JointUrl
			setPhoto = true
		}
	}

	// 没有设置就用默认的第一个
	if !setPhoto {
		if len(se.PhotoInfos) > 0 {
			se.AdPhotoUrl = se.PhotoInfos[0].Url
			se.JointPhotoUrl = se.PhotoInfos[0].JointUrl
		}
	}

	data, err = json.Marshal(se)
	if err == nil {
		ad.Element = string(data)
	} else {
		common.Warningf("logid:%v ScreenModified marshal  error %s", ctx.Value("logid"), err.Error())
	}

	return
}

// 图片像素比例匹配
// 代表图片类型，当前只给开屏用
// 1为android尺寸,2为ios4/4s/pad尺寸,3为ios其他尺寸
func getPhotoType(req *tf.RequestParams) int {
	// 如果是android机，一律返回3
	if req != nil &&
		req.Extra != nil &&
		req.Extra.DeviceOs != nil {
		switch *req.Extra.DeviceOs {
		case "android":
			return 1
		}
	}

	// PixelType =1 | 2 是用来给 apple 用的
	// deviceType 默认空，只对 iphone开头的有效，android机型太复杂，忽略
	var deviceType string
	if req != nil &&
		req.Extra != nil &&
		req.Extra.DeviceType != nil {
		deviceType = strings.ToLower(*req.Extra.DeviceType)
	}

	// 如果是老的iphone 返回1
	if strings.HasPrefix(deviceType, "ipad") ||
		strings.HasPrefix(deviceType, "ipod") ||
		strings.HasPrefix(deviceType, "iphone1") ||
		strings.HasPrefix(deviceType, "iphone2") ||
		strings.HasPrefix(deviceType, "iphone3") ||
		strings.HasPrefix(deviceType, "iphone4") {
		return 2
	}
	// iphone5以上的，返回true来兜底
	if strings.HasPrefix(deviceType, "iphone5") ||
		strings.HasPrefix(deviceType, "iphone6") ||
		strings.HasPrefix(deviceType, "iphone7") ||
		strings.HasPrefix(deviceType, "iphone8") ||
		strings.HasPrefix(deviceType, "iphone9") {
		return 3
	}

	// 默认不对，一律用iphone5以上的尺寸
	return 2
}
