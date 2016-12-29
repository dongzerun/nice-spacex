package defined

var (
	// 获取用户信息版本号 version
	GetVersionSql = "SELECT `version` FROM ad_base.ad_cur_version LIMIT 1"
	// 获取用户信息
	GetUserInfoConstSql = "SELECT immutable_payload, mutable_payload FROM ad_base.ad_user_info_%d WHERE uid=? AND version=? LIMIT 1"
	// 获取nice内部白名单用户
	GetWhiteUsersSql = "SELECT uid FROM tbl_ad_white_list"
	// 获取uid是否关注cuid
	GetUserFollowSql = "select 1 from kk_user_collect where uid=? and cuid=? limit 1"
	//
	GetUserAssetCountSql = "select count(*) from kk_user_asset where uid=? and aid=?"
	// 获取指定广告的过滤条件
	GetFilterByAdSql = "SELECT filter_id, ad_id, filter_name, filter_type,filter_value, create_time,update_time FROM tbl_ad_filters WHERE ad_id = ?"
	// 获取所有有效的广告
	GetAllValidAdsSql = "SELECT ad_id, ad_name, ad_description, ad_area,ad_type,ad_element,status,create_time,update_time FROM tbl_ad WHERE status ='online'"
	// 根据位置获取所有广告
	GetAdsByAreaSql = "SELECT ad_id, ad_name, ad_description, ad_area,ad_type,ad_element,status,create_time,update_time FROM tbl_ad WHERE status in ('online', 'test') and ad_area=?"
	// 获取所有广告位置
	GetAllAreasSql = "SELECT distinct ad_area FROM tbl_ad"
	// 获取所有曝光量配置
	GetExposureSql = "select ad_area, ad_display_limit from ad_base.tbl_ad_exposure"
)
