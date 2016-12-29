package wrapad

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"common"
	tf "spacex_tf/spacex"
	"xad"
	"xutil/hack"
)

func init() {
	// RegisterPostFunc("modify_link_url", PostModifyLink)
}

type LinkElement struct {
	UserProfileUrl string `json:"user_profile_url"` // 用户首页
	Link           string `json:"link"`             // 外连
	Sid            string `json:"sid"`              // 图片SID
	UserType       string `json:"user_type"`        // 用户类型
	UserName       string `json:"user_name"`        // 用户名字
}

func PostModifyLink(ctx context.Context, ad *xad.Ad, req *tf.RequestParams, source string) {
	// 先简单实现，对返回外链类型的广告增加url随机数，使得客户端不做缓存
	// 当前只对 link 外连的做特殊处理
	if ad.Area != "vsfeed_card_3" || ad.Type != "link" {
		return
	}

	//{"user_profile_url":"http:\/\/img08.oneniceapp.com\/upload\/share\/5b9e8d1751c498bcea252129ab374437.jpg",
	//"link":"http:\/\/bsch.serving-sys.com\/BurstingPipe\/adServer.bs?cn=tf&c=20&mc=click&pli=17225410&PluID=0&ord=__TIME__&mb=1",
	//"sid":"84253248813268992",
	// "user_type":"none_user",
	// "user_name":"\u5fc3\u613f\u65e0\u9650\u5927"}

	le := new(LinkElement)
	err := json.Unmarshal(hack.Slice(ad.Element), le)
	if err != nil {
		common.Warningf("postModifyLink Unmarshal newAd.Element error:%s", err.Error())
		return
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano()%10000, 10)

	if strings.Contains(le.Link, "?") {
		le.Link = le.Link + "&nice=" + timestamp
	} else {
		le.Link = le.Link + "?nice=" + timestamp
	}

	var element []byte

	element, err = json.Marshal(le)
	if err != nil {
		common.Warningf("postModifyLink Mmarshal newAd.Element error:%s", err.Error())
		return
	}

	ad.Element = hack.String(element)
	common.Infof("postModifyLink wrap test new url %s", le.Link)
	return
}
