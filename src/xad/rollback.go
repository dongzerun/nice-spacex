package xad

import (
	"common"
	"xad/xplugin"
)

var GlobalCancelChan chan *xplugin.RollBack

func init() {
	if GlobalCancelChan == nil {
		GlobalCancelChan = make(chan *xplugin.RollBack, 2046)
	}

	go CancelLoop()
}

func CancelLoop() {
	for {
		select {
		case rollback := <-GlobalCancelChan:
			if rollback == nil {
				continue
			}
			common.Warningf("logid:%s uid:%d async rollback ad:%d called plugin:%s", rollback.RequestId, rollback.Uid, rollback.Ad, rollback.RbName)
			rollback.Fn()
		}
	}
	common.Warning("CancelLoop quit goroutine")
}
