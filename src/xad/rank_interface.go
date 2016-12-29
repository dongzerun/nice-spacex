package xad

// 所有的rank server 在这里注册
var RankServer map[string]Rank

type Rank interface {
	Rank([]*Ad) []*Ad
}

func RegisterRankServer(name string, r Rank) {
	if RankServer == nil {
		RankServer = make(map[string]Rank)
	}

	if _, ok := RankServer[name]; ok {
		panic("repeat register rank server " + name)
	}
	RankServer[name] = r
}
