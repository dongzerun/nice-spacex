package xad

import (
	"encoding/json"
	simple "github.com/bitly/go-simplejson"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"common"
	"defined"
	"xad/xplugin"
	database "xutil/db"
	"xutil/hack"
)

var (
	// 最大版本号
	MaxVersion []int
)

type LiveParamCondition struct {
	Status     []string `json:"status"`
	Uid        string   `json:"uid"`
	Start_time string   `json:"add_time_start"`
}

// 制定一个获取直播列表的http query参数
// @param uid int64
// @param initBeginTime int64
// @return url.Values,error
func makeLiveListParam(uid int64, initBeginTime int64) (url.Values, error) {
	val := url.Values{}
	condition := LiveParamCondition{
		//只获取直播中和回放状态的直播
		Status: []string{"living", "end"},
		//指定uid
		Uid: strconv.Itoa(int(uid)),
		//设置为当前时间的直播列表
		//Start_time: strconv.Itoa(int(time.Now().Unix())),
		Start_time: strconv.FormatInt(initBeginTime, 10),
	}
	cond, err := json.Marshal(condition)
	if err != nil {
		//fmt.Println("err marshal condition:",err)
		common.Warningf("make livelist http query param error:%s", err.Error())
		return nil, err
	}

	//http://wiki.niceprivate.com/pages/viewpage.action?pageId=7320712
	val.Set("condition", string(cond))
	//http返回值已经按照start_time排序，所以只用获取当前时间的直播列表的第一个
	val.Set("limit", "1")

	val.Set("caller", "data")
	val.Set("request_logid", "123123123")

	return val, nil
}

//获取直播广告当前的信息
//@param uid int64
//@param initBeginTime int64
//@return liveId int64 直播id
//@return lastLiveId int64 最近一次直播ID
//@return liveStatus string 直播当前的状态 forshow online reply
//@return liveLength int 直播时常，如果是foreshow或者online,则为0
func getLiveCurrentInfo(uid int64, initBeginTime int64) (liveId int64, lastLiveId int64, liveStatus string, liveLength int) {
	liveId = 0
	lastLiveId = 0
	liveStatus = "forenotice"
	liveLength = 0

	val, err := makeLiveListParam(uid, initBeginTime)
	if err != nil {
		common.Warningf("make uid:%d livelist http query param error:%s", uid, err.Error())
		return
	}

	//http://wiki.niceprivate.com/pages/viewpage.action?pageId=7320712
	//获取当前时间的直播列表的第一个直播，调用http api
	resp, err := http.PostForm(
		"http://rpc.niceprivate.com/liverpc/LiveApi/getLiveList",
		val,
	)

	defer resp.Body.Close()
	if err != nil {
		common.Warningf("request uid:%d livelist http query error:%s", uid, err.Error())
		return
	} else {
		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			common.Warningf("read uid:%d livelist http response error:%s", uid, err.Error())
			return
		}

		livelist, err := simple.NewJson(resBody)
		if err != nil {
			common.Warningf("make uid:%d livelist simple json error:%s", uid, err.Error())
			return
		}
		//code不为0，说明api调用出错
		code, err := livelist.Get("code").Int()
		if err != nil || code != 0 {
			common.Warningf("get uid:%d livelist simple json code:%d error:%s", uid, code, err.Error())
			return
		}

		count, err := livelist.Get("data").Get("count").Int()
		if err != nil {
			common.Warningf("get uid:%d livelist simple json count:%d error:%s", uid, count, err.Error())
			return
		}
		//如果直播列表为空，代表当前时间用户没有直播，返回一个预上线状态
		if count == 0 {
			common.Infof("uid:%d may be not living yet", uid)
			return
		}
		live, err := livelist.Get("data").Get("data").GetIndex(0).Map()
		if err != nil {
			common.Warningf("get uid:%d livelist simple json live data error:%s", uid, err.Error())
			return
		}
		liveId, err = strconv.ParseInt(live["id"].(string), 10, 64)
		if err != nil {
			common.Warningf("get uid:%d livelist simple json liveId:%d error:%s", uid, liveId, err.Error())
			return
		}
		lastLiveId = liveId
		//http api返回值的end状态就是回放状态
		if live["status"] == "end" {
			liveStatus = "reply"
			//计算直播时间
			end_time, err := strconv.Atoi(live["end_time"].(string))
			if err != nil {
				common.Warningf("get uid:%d livelist simple json end_time:%d error:%s", uid, end_time, err.Error())
				return
			}
			start_time, err := strconv.Atoi(live["start_time"].(string))
			if err != nil {
				common.Warningf("get uid:%d livelist simple json end_time:%d error:%s", uid, start_time, err.Error())
				return
			}
			liveLength = end_time - start_time
		} else {
			liveStatus = "online"
			liveLength = 0
		}
	}

	return
}

func init() {
	rand.Seed(time.Now().UnixNano())

	MaxVersion = make([]int, 10)
	for i, _ := range MaxVersion {
		MaxVersion[i] = math.MaxInt32
	}
}

// 根据广告ID获取对应的条件信息
func GetFilterByAdID(id int) ([]*Filter, error) {
	db := database.AdGetter.P.GetConn()
	if db == nil {
		return nil, common.MySQLPoolTimeOutError
	}
	defer database.AdGetter.P.Release(db)

	rows, err := db.Query(defined.GetFilterByAdSql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fs := make([]*Filter, 0, 10)
	for rows.Next() {
		f := &Filter{}
		e := rows.Scan(&f.FId, &f.AId, &f.Name, &f.Type, &f.Value, &f.CreateTime, &f.UpdateTime)
		if e != nil {
			return nil, e
		}
		fs = append(fs, f)
	}
	return fs, nil
}

// 获取当前所有可用的广告信息
func GetAllValidAds() ([]*Ad, error) {
	db := database.AdGetter.P.GetConn()
	if db == nil {
		return nil, common.MySQLPoolTimeOutError
	}
	defer database.AdGetter.P.Release(db)

	rows, err := db.Query(defined.GetAllValidAdsSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ads := make([]*Ad, 0, 100)
	for rows.Next() {
		ad := &Ad{}
		e := rows.Scan(&ad.Id, &ad.Name, &ad.Description, &ad.Area, &ad.Type, &ad.Element, &ad.Status, &ad.CreateTime, &ad.UpdateTime)
		if e != nil {
			return nil, e
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

// 获取指定位置 的广告
func GetAdsByArea(area string) ([]*Ad, error) {
	db := database.AdGetter.P.GetConn()
	if db == nil {
		return nil, common.MySQLPoolTimeOutError
	}
	defer database.AdGetter.P.Release(db)

	rows, err := db.Query(defined.GetAdsByAreaSql, area)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ads := make([]*Ad, 0, 100)
	for rows.Next() {
		ad := &Ad{
			WhiteList: make([]int, 0),
		}
		e := rows.Scan(&ad.Id, &ad.Name, &ad.Description, &ad.Area, &ad.Type, &ad.Element, &ad.Status, &ad.CreateTime, &ad.UpdateTime)
		if e != nil {
			return nil, e
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

// 获取所有广告的位置
func GetAllAreas() []string {
	areas := make([]string, 0, 3)
	db := database.AdGetter.P.GetConn()
	if db == nil {
		return areas
	}
	defer database.AdGetter.P.Release(db)

	rows, err := db.Query(defined.GetAllAreasSql)
	if err != nil {
		return areas
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		e := rows.Scan(&name)
		if e != nil {
			return areas
		}
		areas = append(areas, name)
	}
	return areas
}

// 定期去更新广告信息
func LoopReloadAds() {

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			// 重新读取所有 area
			AdAreas = GetAllAreas()
			// 重新读取所有area下面的广告
			GetAds()
		}
	}
}

func GetAds() {
	defer func() {
		// 捕获 当前goroutine的error
		if err := recover(); err != nil {
			buf := make([]byte, defined.STACKSIZE)
			buf = buf[:runtime.Stack(buf, false)]
			common.Errorf("Spacex GetAds panic %s", string(buf))
		}
	}()
	for _, area := range AdAreas {

		ads, err := GetAdsByArea(area)
		if err != nil {
			common.Warning("AdGetter GetAdsByArea error ", err.Error())
		}

		common.Infof("GetAds start get area:%s, length:%d", area, len(ads))

		for i, _ := range ads {
			updateFilter2(ads[i])
			if (ads[i].Area == "vsfeed_card_3" && ads[i].Type == "live") ||
				(ads[i].Area == "live_index" && ads[i].Type == "live_fixed") {
				elements, err := simple.NewJson([]byte(ads[i].Element))
				if err != nil {
					common.Warningf("make ad %s %s live elements simple json error:%s", ads[i].Area, ads[i].Type, err.Error())
					continue
				}
				//获取到有直播广告的用户id,vsfeed_card_3或者liveindex
				uid, err := elements.Get("live_uid").Int64()
				if err != nil {
					common.Warningf("get ad %s %s uid error:%s", ads[i].Area, ads[i].Type, err.Error())
					continue
				}
				initBeginTime, err := elements.Get("live_initial_begin").Int64()
				if err != nil {
					//live_index里面没有live_initial_begin,
					//取广告生效时间的起点时间
					//注：只能查询到initBeginTime以后开播的广告信息
					for _, f := range ads[i].Filters {
						if f.Name == "valid_time" {
							initBeginTime = f.Matcher.(*xplugin.ValidTimeRange).Time[0]
							break
						}
					}
				}
				common.Infof("%s %s initBeginTime:%d", ads[i].Area, ads[i].Type, initBeginTime)
				//获取该用户的直播广告状态、id、直播时长
				liveId, lastLiveId, status, liveLength := getLiveCurrentInfo(uid, initBeginTime)
				//设置直播信息到ad_element，供返回广告时直接获取ad_elements
				elements.Set("live_id", liveId)
				elements.Set("last_lid", lastLiveId)
				elements.Set("live_status", status)
				elements.Set("last_len", liveLength)
				elementsBytes, err := elements.MarshalJSON()
				if err != nil {
					common.Warningf("marshal ad %s %s live elements error:%s", ads[i].Area, ads[i].Type, err.Error())
				}
				//ad_manager接管该广告，spacex返回广告时，总是从ad_manager中获取广告基本信息
				ads[i].Element = string(elementsBytes)
			}
			common.Info("get ad filter size: ", len(ads[i].Filters), " detail: ", ads[i])
			for idx, _ := range ads[i].Filters {
				data, err := json.Marshal(ads[i].Filters[idx])
				if err != nil {
					continue
				}
				common.Info("ad filter ", idx, "\t", hack.String(data))
			}
		}

		// 只有vsfeed_card_3才有运营广告，需要区分
		// 其它默认都是ad即可
		for _, ad := range ads {
			if strings.HasPrefix(ad.Type, "op_") {
				ad.OpOrAd = "op"
			} else {
				ad.OpOrAd = "ad"
			}
			// add random RankWeight
			ad.RankWeight = rand.Intn(100)
		}

		// Rank 操作留到AdM.Set
		// RankAds(ads)

		if len(ads) == 0 {
			common.Warningf("GetAds %s ads length == 0 , remove %s from ADM", area, area)
			AdM.Del(area) // 删除 为空的广告列表
			continue
		}

		AdM.Set(area, ads)
	}
}
