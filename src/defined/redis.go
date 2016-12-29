package defined

import (
// "fmt"
)

var (
	// redis用户信息 todo废弃
	KeyUICacheFormat = "userinfo_%d"
	// redis用户信息  snappy压缩
	KeyUICacheSnappyFormat = "userinfosnappy_%d"
	// redis用户信息  lz4压缩
	KeyUICacheLz4Format = "userinfolz4_%d"
	// 提定用户与广告展示次数计数, 第一个参数是adid, 第二个是ad status 第三个是uid或did
	KeyAdUserCntFormat = "adusercnt_%d_%s_%v"
	// 提定用户与广告展示次数计数v2版本
	KeyAdUserCntV2Format = "adusercntv2_%d_%v"
	// %d 是UID,value是 AdId
	// %s 是method,即广告的ad_area位置
	KeyUidLastAdFormat = "uidad_%d_%s"
	// 只cache tag link的广告, 是在feed card里命中的
	KeyTagLinkADCacheFormat = "aduserad:%v_%v"
	// 反馈用户屏蔽指定广告前一个是uid,后一个是adid
	KeyBlockActionFormat = "adblock_%d_%d"
	// show_cache 中保存了用户打出标签的缓存
	KeyTagUserCacheFormat = "nice_tag_tag_user_tag_pub:%d"
	// tag detail 的key
	KeyTagPublicTagFormat = "%d_exists"
	// 为了防止抖动，连续5内内的对同一广告请求只计一次,第一个是adid,第二个是uid
	KeyTimeJitterFormat = "timejitter_%d_%d"
	// 某人的关注列表 参数是 uid
	KeyUserFollowingFormat = "nice_user_relationship_following:%d"
	// 曝光量限制参数是 ad_id, ad_id 状态
	KeyAdExposureFormat = "ad_exposure_%d_%s"
)
