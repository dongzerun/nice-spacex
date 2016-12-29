package xad

import (
	"runtime"

	"golang.org/x/net/context"

	"common"
	"config"
	"defined"
	"xad/xplugin"
	"xuser"
)

type Executor struct {
}

// MultiGo 计划采用并发获取模式，对于多广告获取时帮助很大，并发限制在5就可以了
// 匹配结束有三个标志：
// 1. 匹配单个广告，有结果就返回
// 2. 遍历完所有候选广告
// 3. 到达了服务约定的超时时间
//
// 有几个问题没有想好：
// 1. 对于非白名单用户广告有优先级之分，优先展示 online 广告，同时在线运营卡片优先级低于广告卡片
// 排序问题，以前是顺序匹配，所以在获取AD列表时rank好序就可以，现在并发去匹配问题就多了
// 2. 广告匹配时产生了状态，如果没匹配需要做状态的回滚，最基本的就是 max display num 次数
// 3. 对于单个广告，匹配到后，要取消其它并发操作，做到高扩展，模块划分要清晰
//
// 另外何时退出并发:
// 1. 所有的都以服务端传来的context超时为最终超时
// 2. 只需要单个广告:
//   1)如果channel出来的是线上,并且是AD广告卡片，并且非预上线用户，直接取消其它操作，回滚，返回该AD
//   2)对于预上线用户，只要有一个命中，就返回一个AD，回滚其它
// 3. 对于多个广告的，遍历等待所有有效广告，返回ad slice，回滚其它
//   多广告目前只存在于发现页，其它主要的 vsfeed_card_3 都是单张广告
//
func (e *Executor) MultiGo(ctx context.Context, ui *xuser.UserInfo, adverts []*Ad, isMore bool) ([]*Ad, error) {
	// 获取到的广告
	ads := make([]*Ad, 0, 5)
	// 回滚资源池
	rbs := make([]*xplugin.RollBack, 0, 5)
	// 广告候选集 channel
	candidates := make(chan *Ad, config.GlobalConfig.MiscConfig.ConcurrentMatch)
	// 资源回滚 channel
	rbChannel := make(chan *xplugin.RollBack, len(adverts))
	// 基于父 ctx，派生出带有取消函数的新 context，完全继承自父ctx
	derivedCtx, cancelFunc := context.WithCancel(ctx)
	// 待执行次数
	total := len(adverts)
	// 已执行次数
	finished := 0
	// 已执行 channel
	finChannel := make(chan struct{}, total)

	// 通过channel 控制并发，填充数据，空结构体只占 1 byte, 配置文件默认是5个ad
	concurrent := make(chan struct{}, config.GlobalConfig.MiscConfig.ConcurrentMatch)
	for i := 0; i < config.GlobalConfig.MiscConfig.ConcurrentMatch; i++ {
		concurrent <- struct{}{}
	}

	// 打到后台并发执行，这块不会有问题，不用担心panic
	// recover 逻辑扔到 enterGo
	go func() {
		common.Infof("logid:%v uid:%d MultiGo area adverts length:%d", ctx.Value("logid"), ui.Uid, len(adverts))
		for _, ad := range adverts {
			<-concurrent
			release := func() {
				common.Infof("logid:%v uid:%d MultiGo release concurrent lock", ctx.Value("logid"), ui.Uid)
				concurrent <- struct{}{}
				finChannel <- struct{}{}
			}

			// 单独开启 goroutine 去匹配
			go e.enterGo(derivedCtx, release, ui, rbChannel, ad, candidates)
		}
	}()

eventLoop:
	for {
		select {
		// 处理匹配的广告
		case ad := <-candidates:
			common.Infof("logid:%v MultiGo uid:%d receive ad:%v", ctx.Value("logid"), ui.Uid, ad)
			ads = append(ads, ad)
			// 只需要匹配单个广告的用户
			if !isMore {
				// 调用cancel函数，取消其它并行执行分支
				// 并直接退出大循环
				cancelFunc()
			}
		case rb := <-rbChannel:
			common.Infof("logid:%v MultiGo uid:%d receive callback:%v", ctx.Value("logid"), ui.Uid, rb)
			rbs = append(rbs, rb)
		// 如果全匹配完了，也没有广告，那么要退出循环
		case <-finChannel:
			finished++
			if finished == total {
				break eventLoop
			}
		// 如果收到退出信号，退出循环
		case <-derivedCtx.Done():
			common.Warningf("logid:%v MultiGo uid:%d receive derived Cancel", ctx.Value("logid"), ui.Uid)
			break eventLoop
		// 如果收到服务整体超时，退出循环
		case <-ctx.Done():
			common.Warningf("logid:%v MultiGo uid:%d receive father Cancel", ctx.Value("logid"), ui.Uid)
			break eventLoop
		}
	}

	// 没有匹配广告，直接返回 ErrMissingMatchAd
	// 同时也要回滚未使用的广告
	if len(ads) == 0 {
		// 回滚所有 rollback callback
		e.asyncRollBack(ctx, ads, rbs)
		return nil, defined.ErrMissingMatchAd
	}

	// 只需匹配单个广告
	if !isMore {
		e.asyncRollBack(ctx, ads[:1], rbs)
		return ads[:1], nil

	}

	// 匹配多个广告
	e.asyncRollBack(ctx, ads, rbs)
	return ads, nil
}

// release 是释放并发资源的闭包，最后由defer执行
// ui 是 UserInfo 用户相关信息
// rb 是回滚的channel，由xplugin往里写数据，无论ab是否最终匹配，都可能会写rb
// ad 是待匹配的广告
// candidates 是候选广告集channel，如果该 ad命中了，要写到这个候选集里面
func (e *Executor) enterGo(ctx context.Context, release func(), ui *xuser.UserInfo, rb chan *xplugin.RollBack, ad *Ad, candidates chan *Ad) {
	defer func() {
		// 最后首先释放并发控制的资源
		release()
		// 捕获 当前goroutine的error
		if err := recover(); err != nil {
			buf := make([]byte, defined.STACKSIZE)
			buf = buf[:runtime.Stack(buf, false)]
			common.Errorf("logid:%v executor enterGo panic ", ctx.Value("logid"), string(buf))
		}
	}()

	select {
	case <-ctx.Done():
		common.Warningf("logid:%v enterGo Ad:%d timeout or canceled, directly return", ctx.Value("logid"), ad.Id)
		return
	default:
	}

	common.Debugf("logid:%v Match ad:%v", ctx.Value("logid"), ad)

	if ok := ad.match(ctx, ui, rb); ok {
		candidates <- ad
	}
}

// usedAds 返回给客户端的广告集
// rbs 所有注册进来的回滚回调函数
// 把除 usedAds 广告以外的所有 rollback 扔到异步执行 GlobalCancelChan
func (e *Executor) asyncRollBack(ctx context.Context, usedAds []*Ad, rbs []*xplugin.RollBack) {
	for _, rb := range rbs {
		used := false
		for _, ad := range usedAds {
			if rb.Ad == ad.Id {
				// 命中这个广告，那么不需要回滚
				// 退出里层 for 循环
				used = true
				break
			}
		}

		// 如果rb对应的广告未被使用，那么回滚
		if !used {
			select {
			case GlobalCancelChan <- rb:
			default:
				// 下面的日志理论上一条都不应用，有则说明异步消费 callback慢了
				common.Warningf("logid:%v MultiGo rollback rejected ad:%d", ctx.Value("logid"), rb.Ad)
			}
		}
	}
}
