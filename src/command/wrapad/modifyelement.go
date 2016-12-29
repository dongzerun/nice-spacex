package wrapad

import (
	"strconv"

	simple "github.com/bitly/go-simplejson"
	"golang.org/x/net/context"

	"common"
	tf "spacex_tf/spacex"
	"xad"
	"xutil/hack"
)

func init() {
	RegisterPostFunc("modify_ad_element", PostModifyElement)
}

func PostModifyElement(ctx context.Context, ad *xad.Ad, req *tf.RequestParams, source string) {
	sj, err := simple.NewJson(hack.Slice(ad.Element))
	if err != nil {
		common.Warningf("postModifyElement NewJson failed %s", err.Error())
		return
	}

	sj.Set("req_uid", strconv.FormatInt(*req.UID, 10))

	data, err := sj.MarshalJSON()
	if err != nil {
		common.Warningf("postModifyElement MarshalJSON failed %s", err.Error())
		return
	}

	ad.Element = hack.String(data)
}
