package defined

// iphone1,1   iphone
// iphone1,2   iphone 3g
// iphone2,1   iphone 3gs
// iphone3,1   iphone 4 (gsm)
// iphone3,3   iphone 4 (cdma)
// iphone4,1   iphone 4s
// iphone5,1   iphone 5 (a1428)
// iphone5,2   iphone 5 (a1429)
// iphone5,3   iphone 5c (a1456/a1532)
// iphone5,4   iphone 5c (a1507/a1516/a1529)
// iphone6,1   iphone 5s (a1433/a1453)
// iphone6,2   iphone 5s (a1457/a1518/a1530)
// iphone7,1   iphone 6 plus
// iphone7,2   iphone 6
// iphone8,1   iphone 6s
// iphone8,2   iphone 6s plus
// ipad1,1 ipad
// ipad2,1 ipad 2 (wi-fi)
// ipad2,2 ipad 2 (gsm)
// ipad2,3 ipad 2 (cdma)
// ipad2,4 ipad 2 (wi-fi, revised)
// ipad2,5 ipad mini (wi-fi)
// ipad2,6 ipad mini (a1454)
// ipad2,7 ipad mini (a1455)
// ipad3,1 ipad (3rd gen, wi-fi)
// ipad3,2 ipad (3rd gen, wi-fi+lte verizon)
// ipad3,3 ipad (3rd gen, wi-fi+lte at&t)
// ipad3,4 ipad (4th gen, wi-fi)
// ipad3,5 ipad (4th gen, a1459)
// ipad3,6 ipad (4th gen, a1460)
// ipad4,1 ipad air (wi-fi)
// ipad4,2 ipad air (wi-fi+lte)
// ipad4,3 ipad air (rev)
// ipad4,4 ipad mini 2 (wi-fi)
// ipad4,5 ipad mini 2 (wi-fi+lte)
// ipad4,6 ipad mini 2 (rev)
// ipad4,7 ipad mini 3 (wi-fi)
// ipad4,8 ipad mini 3 (a1600)
// ipad4,9 ipad mini 3 (a1601)
// ipad5,1 ipad mini 4 (wi-fi)
// ipad5,2 ipad mini 4 (wi-fi+lte)
// ipad5,3 ipad air 2 (wi-fi)
// ipad5,4 ipad air 2 (wi-fi+lte)
// ipad6,7 ipad pro (wi-fi)
// ipad6,8 ipad pro (wi-fi+lte)
// ipod1,1 ipod touch
// ipod2,1 ipod touch (2nd gen)
// ipod3,1 ipod touch (3rd gen)
// ipod4,1 ipod touch (4th gen)
// ipod5,1 ipod touch (5th gen)
// ipod7,1 ipod touch (6th gen)

//公司内部测试device_id
var preOnlineDevices = []string{"cba52ca916c6144009bc8cca5ca075f58",
	"89096582033abbf59d6c278077c48752",
	"3ca55aee14c64adfc1f837244d3f455f",
	"ba52ca916c6144009bc8cca5ca075f58",
	"dbbdec1950c445b8a686533174e2126b",
	"c5423d8b79d7b615a603ccdf8fac104a",
	"4c31364c6768e999788c306371c9cb72",
	"cf754a277faf3c0e4452dec5e7746d93",
	"dda18321797a8ade312fed7d8f3ff0f7",
	"c5423d8b79d7b615a603ccdf8fac104a",
}

func IsDeviceInPreOnline(id string) bool {
	for i, _ := range preOnlineDevices {
		if id == preOnlineDevices[i] {
			return true
		}
	}
	return false
}
