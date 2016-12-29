package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"time"
)

var (
	GlobalConfig Config
)

type Config struct {
	SConfig          ServerConfig           `toml:"server"`       // 系充配置
	LLConfig         LongLatConfig          `toml:"longlat_conf"` // 经纬度配置
	LogConfig        LogConfig              `toml:"golog"`        // 日志配置
	AdDBConfig       DBconfig               `toml:"addb"`         // 广告数据库配置
	UserDBConfig     DBconfig               `toml:"userdb"`       // user 数据库配置
	CoreShowConfig   DBconfig               `toml:"coreshow"`     // 关注列表配置
	UserTagConfig    DBconfig               `toml:"usertag"`      // 用户发标签配置
	UserPasterConfig DBconfig               `toml:"userpaster"`   // 用户发贴纸配置
	UserAssetConfig  DBconfig               `toml:"userasset"`
	InfluxConfig     InfluxConfig           `toml:"influxdb"`  // influxdb配置
	CacheConfig      map[string]Redisconfig `toml:"cache"`     // redis cache配置
	ShowCacheConfig  map[string]Redisconfig `toml:"showcache"` // show cache 存放标签
	SocialConfig     map[string]Redisconfig `toml:"social"`    // social关注关系集群
	MiscConfig       Misc                   `toml:"misc"`      // 杂项配置
}

type LongLatConfig struct {
	Longlat2Code  string `toml:"longlat2code"`
	Code2City     string `toml:"code2city"`
	Code2Province string `toml:"code2province"`
}

type ServerConfig struct {
	MaxCpu         int `toml:"max_cpu"`
	ThriftPort     int `toml:"thrift_port"`
	PprofPort      int `toml:"pprof_port"`
	RestPort       int `toml:"rest_port"`
	DefaultTimeout int `toml:"timeout"`
}

type LogConfig struct {
	Level      string `toml:"level"`
	Console    int    `toml:"console"`
	Dir        string `toml:"dir"`
	Filename   string `toml:"filename"`
	ReserveNum int    `toml:"reserve_num"`
	Suffix     string `toml:"suffix"`
	Colorfull  int    `toml:"colorfull"`
}

type DBconfig struct {
	Usr      string `toml:"user"`
	Pwd      string `toml:"pwd"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DBname   string `toml:"db_name"`
	MaxIdle  int    `toml:"max_idle"`
	MaxOpen  int    `toml:"max_open"`
	PoolSize int    `toml:"pool_size"`
}

// idleTimeout, connectTimeout, readTimeout, writeTimeout time.Duration, poolSize
type Redisconfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	PoolSize int    `toml:"pool_size"`
	MaxIdle  int    `toml:"max_idle"`

	IdleTimeout    time.Duration `toml:"idle_timeout"`
	ConnectTimeout time.Duration `toml:"connect_timeout"`
	ReadTimeout    time.Duration `toml:"read_timeout"`
	WriteTimeout   time.Duration `toml:"write_timeout"`
}

type InfluxConfig struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	DB         string `toml:"db"`
	User       string `toml:"user"`
	Pwd        string `toml:"pwd"`
	BufferSize int    `toml:"buffer_size"`
}

type Misc struct {
	UserCompressed  string `toml:usercompressed`
	ConcurrentMatch int    `toml:concurrentmatch`
}

func InitConfig(filename string) {
	GlobalConfig = NewTomlConfig(filename)
}

// 获取config
func NewTomlConfig(filename string) (conf Config) {

	var (
		data []byte
		err  error
	)

	data, err = ioutil.ReadFile(filename)

	if err != nil {
		panic("read configuration file failed " + err.Error())
	}

	if _, err = toml.Decode(string(data), &conf); err != nil {
		panic("toml decode failed " + err.Error())
	}
	return
}
