package xuser

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"common"
	"defined"
	database "xutil/db"
)

// 每分钟去定时获取最新的数据版本号
func LoopVersion() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			GetInfoVersion()
		}
	}

	common.Warning("Spacex User Info LoopVersion quit goroutine")
}

func GetInfoVersion() {
	db := database.InfoGetter.P.GetConn()
	if db == nil {
		return
		//return nil, common.MySQLPoolTimeOutError
	}
	defer database.InfoGetter.P.Release(db)

	rows, err := db.Query(defined.GetVersionSql)
	if err != nil {
		common.Warning("GetInfoVersion query error ", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		e := rows.Scan(&UserVersion)
		if e != nil {
			common.Warning("GetInfoVersion Scan error ", e.Error())
			return
		}
	}
}

// immutable string, mutable string, err error
func GetUserInfo(ctx context.Context, uid int64) (string, string, error) {
	select {
	case <-ctx.Done():
		common.Warningf("logid:%v GetUserInfo timeout or canceled, directly return", ctx.Value("logid"))
		return "", "", defined.ErrContextTimeout
	default:
	}

	start := time.Now().UnixNano()
	defer common.Infof("logid:%v GetUserInfo DB consumed time:%d", ctx.Value("logid"), time.Now().UnixNano()-start)
	if UserVersion < 0 {
		GetInfoVersion()
	}

	db := database.InfoGetter.P.GetConn()
	if db == nil {
		return "", "", common.MySQLPoolTimeOutError
	}
	defer database.InfoGetter.P.Release(db)

	rows, err := db.Query(GenGetUserSql(uid), uid, UserVersion)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	common.Infof("logid:%v GetUserInfo from DB:%d SQL:%s UserVersion:%d", ctx.Value("logid"), uid, GenGetUserSql(uid), UserVersion)

	var (
		immutable string
		mutable   string
	)

	for rows.Next() {
		e := rows.Scan(&immutable, &mutable)
		if e != nil {
			return "", "", e
		}
	}

	return immutable, mutable, nil
}

func GenGetUserSql(uid int64) string {
	return fmt.Sprintf(defined.GetUserInfoConstSql, uid%100)
}
