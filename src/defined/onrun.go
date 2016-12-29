package defined

// 全局初始化注册列表
var ServerOnRun map[string]func()

func init() {
	if ServerOnRun == nil {
		ServerOnRun = make(map[string]func())
	}
}

// 模块在这里注册
func RegisterOnRun(name string, fn func()) {
	_, exists := ServerOnRun[name]
	if exists {
		panic(name + " ServerOnRun duplicate Register")
	}
	ServerOnRun[name] = fn
}
