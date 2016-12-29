package defined

import (
	tf "spacex_tf/spacex"
)

const (
	// 返回正常
	StatusOK int32 = iota
	// 服务内部超时
	StatusTimeout
	// 没有找到匹配的method
	StatusNotFoundMethod
	// 没有用户信息
	StatusMissUserInfo
	// 没有匹配到ad
	StatusMissMatchAd
	// 未知错误
	StatusUnknownError
	// 贴纸包错误
	StatusPackageResultError
	// 传入uid为空错误
	StatusReqUidNull
	// ad id 不存在，或无效
	StatusAdIdNonExists
	// action不存在或非法
	StatusActionEmptyOrIllegal
)

var (
	RespOK                   = newResponse(StatusOK, "success")
	RespTimeout              = newResponse(StatusTimeout, "timeout")
	RespNotFoundMethod       = newResponse(StatusNotFoundMethod, "rpc method unvaliad")
	RespMissUserInfo         = newResponse(StatusMissUserInfo, "missing userinfo")
	RespMissMatchAd          = newResponse(StatusMissMatchAd, "not match ad")
	RespUnknownError         = newResponse(StatusUnknownError, "unknonw error")
	RespPackageResultError   = newResponse(StatusUnknownError, "unknown error")
	RespReqUidNull           = newResponse(StatusReqUidNull, "request uid null")
	RespAdIdNonExists        = newResponse(StatusAdIdNonExists, "get ad detail failed")
	RespActionEmptyOrIllegal = newResponse(StatusActionEmptyOrIllegal, "request action empty or illegal")
)

func newResponse(status int32, error_string string) *tf.Response {
	return &tf.Response{
		Status:      status,
		ErrorString: &error_string,
	}
}

func GetResponse(status int32) (res *tf.Response) {
	switch status {
	case StatusOK:
		return newResponse(StatusOK, "success")
	case StatusTimeout:
		return RespTimeout
	case StatusNotFoundMethod:
		return RespNotFoundMethod
	case StatusMissUserInfo:
		return RespMissUserInfo
	case StatusMissMatchAd:
		return RespMissMatchAd
	case StatusUnknownError:
		return RespUnknownError
	case StatusPackageResultError:
		return RespPackageResultError
	case StatusReqUidNull:
		return RespReqUidNull
	case StatusAdIdNonExists:
		return RespAdIdNonExists
	case StatusActionEmptyOrIllegal:
		return RespActionEmptyOrIllegal
	}
	return RespUnknownError
}
