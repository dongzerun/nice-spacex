package db

import (
	"common"

	"config"
	// "defined"
)

var (

	// 全局用户数据库 单例
	InfoGetter *DBSlave

	// 全局广告数据库
	AdGetter *DBSlave

	// coreshow 关注列表数据库
	CoreShow *DBSlave

	// usershow 用户发标签
	UserTag *DBSlave

	// userpaster 用户发贴纸
	UserPaster *DBSlave

	// UserAsset 用户关注标签
	UserAsset *DBSlave
)

// func init() {
// 	defined.RegisterOnRun("db", func() {
// 		common.Info("Server On Run db.InitDB")
// 		InitDB()
// 	})
// }

func InitDB() {

	if InfoGetter == nil {
		InfoGetter = NewDBSlave(config.GlobalConfig.UserDBConfig)
	}

	if AdGetter == nil {
		AdGetter = NewDBSlave(config.GlobalConfig.AdDBConfig)
	}

	if CoreShow == nil {
		CoreShow = NewDBSlave(config.GlobalConfig.CoreShowConfig)
	}

	if UserTag == nil {
		UserTag = NewDBSlave(config.GlobalConfig.UserTagConfig)
	}

	if UserPaster == nil {
		UserPaster = NewDBSlave(config.GlobalConfig.UserPasterConfig)
	}

	if UserAsset == nil {
		UserAsset = NewDBSlave(config.GlobalConfig.UserAssetConfig)
	}

}

type DBSlave struct {
	P *common.DBPool
}

func NewDBSlave(c config.DBconfig) *DBSlave {
	ds := &DBSlave{}
	p, err := common.NewDBPool(
		c.Usr,
		c.Pwd,
		c.Host,
		c.Port,
		c.DBname,
		c.MaxIdle,
		c.MaxOpen,
		c.PoolSize)

	if err != nil {
		panic("dbpool failed " + err.Error())
	}

	ds.P = p

	return ds
}
