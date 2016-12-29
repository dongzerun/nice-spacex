package xuser

import (
	"sync"
	"time"

	"common"
	"defined"
	database "xutil/db"
)

func init() {
	defined.RegisterOnRun("preonline", func() {
		common.Info("Server On Run xuser.InitWL")
		InitPreOnlineList()
	})
}

type PreOnlineList struct {
	sync.RWMutex
	List map[int64]struct{}
}

var (
	adPreOnline *PreOnlineList
)

func InitPreOnlineList() {
	if adPreOnline == nil {
		adPreOnline = NewAdPreOnlineList()
	}
}

func NewAdPreOnlineList() *PreOnlineList {
	po := &PreOnlineList{}
	po.Reload()
	go po.loopReload()
	return po
}

func (po *PreOnlineList) Reload() {
	l := getPreOnlineList()

	po.Lock()
	defer po.Unlock()
	common.Info("reload preonline list size: ", len(l), l)
	po.List = l
}

func (po *PreOnlineList) loopReload() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			po.Reload()
		}
	}
	common.Warning("preonline quit loop Reload goroutine")
}

// 是否在白名单中
func IsInPreOnlineList(uid int64) bool {
	adPreOnline.RLock()
	defer adPreOnline.RUnlock()

	if _, ok := adPreOnline.List[uid]; ok {
		return true
	}
	return false
}

// 获取白名单用户
func getPreOnlineList() (wusers map[int64]struct{}) {

	wusers = make(map[int64]struct{})

	db := database.AdGetter.P.GetConn()

	if db == nil {
		common.Error("Get preonline list db connect error ", common.MySQLPoolTimeOutError)
		return
	}
	defer database.AdGetter.P.Release(db)

	rows, err := db.Query(defined.GetWhiteUsersSql)

	if err != nil {
		return
	}
	defer rows.Close()

	var (
		uid int64
	)

	for rows.Next() {
		e := rows.Scan(&uid)
		if e != nil {
			return
		}
		wusers[uid] = struct{}{}
	}
	return
}
