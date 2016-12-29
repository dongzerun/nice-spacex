package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"common"
	"config"
	"defined"
	"runtime"
	"xutil/db"
)

var (
	config_file = flag.String("config_file", "", "spacex config file must be toml file")
)

func init() {
	defined.RegisterOnRun("sys", func() {
		common.Info("Server On Run init_sys")
		init_sys()
	})

	defined.RegisterOnRun("influxdb", func() {
		common.Info("Server On Run init_influxdb")
		init_influxdb()
	})
}

func main() {
	flag.Parse()
	//初始化全局配置
	config.InitConfig(*config_file)

	//初始化日志模块，这个必须单独拿出来先初始化
	init_logger()

	common.Info("Server On Run db.InitDB")
	db.InitDB()

	common.Info("init all ServerOnRun")

	for _, fn := range defined.ServerOnRun {
		fn()
	}

	// 启动 http api 服务
	common.Info("start api server")
	StartApiServer(config.GlobalConfig.SConfig.RestPort)

	// 启动eventlisteser
	// startEventListener()

	// 启动 thrift rpc服务
	common.Info("start ad server")
	StartAdServer(config.GlobalConfig.SConfig.ThriftPort)

}

// 初始化sys 配置
func init_sys() {
	// 设置cores
	if config.GlobalConfig.SConfig.MaxCpu == 0 {
		config.GlobalConfig.SConfig.MaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(config.GlobalConfig.SConfig.MaxCpu)

	// 启动ppref 检测
	if config.GlobalConfig.SConfig.PprofPort != 0 {
		go func() {
			common.Info(http.ListenAndServe(fmt.Sprintf(":%d", config.GlobalConfig.SConfig.PprofPort), nil))
		}()
	}
}

// 初始化日志
func init_logger() {
	// 初始化logger
	common.InitLooger(
		config.GlobalConfig.LogConfig.Level,
		config.GlobalConfig.LogConfig.Dir,
		config.GlobalConfig.LogConfig.Filename,
		config.GlobalConfig.LogConfig.ReserveNum,
		config.GlobalConfig.LogConfig.Suffix,
		config.GlobalConfig.LogConfig.Console,
		config.GlobalConfig.LogConfig.Colorfull)
}

func init_influxdb() {
	err := common.InitHandler(config.GlobalConfig.InfluxConfig.Host,
		config.GlobalConfig.InfluxConfig.Port,
		config.GlobalConfig.InfluxConfig.DB,
		config.GlobalConfig.InfluxConfig.User,
		config.GlobalConfig.InfluxConfig.Pwd,
		config.GlobalConfig.InfluxConfig.BufferSize)

	if err != nil {
		common.Warning("spacex influxdb handler init failed ", err.Error())
	}
}
