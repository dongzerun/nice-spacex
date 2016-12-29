package loc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"common"
	"config"
	"defined"
)

type LLLoc interface {
	GetLoc(args ...string) string
	Reload()
}

var (
	_           LLLoc = (*LonglatLoc)(nil) // 纬经度匹配类
	GlobalLLLoc LLLoc                      // 纬经度匹配类
)

const (
	Start_x = 73550
	Start_y = 18100
	Step_x  = 24
	Step_y  = 18

	Unknwon  = "unknown"
	City     = "city"
	Province = "province"
)

func init() {
	defined.RegisterOnRun("loc", func() {
		common.Info("Server On Run loc.InitLoc")
		InitLoc()
	})
}

type LLMaps struct {
	codeMap     map[string]string
	provinceMap map[string]string
	cityMap     map[string]string
}

type LonglatLoc struct {
	sync.RWMutex
	longlat2CodeFile string
	code2City        string
	code2Province    string
	curLLMaps        *LLMaps
}

func NewLongLatLoc() LLLoc {
	llLoc := &LonglatLoc{
		longlat2CodeFile: config.GlobalConfig.LLConfig.Longlat2Code,
		code2City:        config.GlobalConfig.LLConfig.Code2City,
		code2Province:    config.GlobalConfig.LLConfig.Code2Province,
	}
	llLoc.Reload()
	go llLoc.ReloadLoop()
	return llLoc
}

func (this *LonglatLoc) ReloadLoop() {
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			this.Reload()
		}
	}
}

func (this *LonglatLoc) Reload() {
	var (
		com map[string]string
		pm  map[string]string
		cim map[string]string
		err error
	)
	com, err = buildCodeMap(this.longlat2CodeFile)
	if err != nil {
		common.Error("build code map error")
		return
	}
	common.Infof("build longlat code map size: %d", len(com))

	pm, err = buildProvince(this.code2Province)
	if err != nil {
		common.Error("build code to province error")
		return
	}
	common.Infof("build code province map size: %d", len(pm))

	cim, err = buildCity(this.code2City)
	if err != nil {
		common.Error("build code to city error")
		return
	}
	common.Infof("build code city map size: %d", len(cim))

	llMaps := &LLMaps{
		codeMap:     com,
		provinceMap: pm,
		cityMap:     cim,
	}
	this.Lock()
	defer this.Unlock()
	this.curLLMaps = llMaps
}

func GetLoc(args ...string) string {
	return GlobalLLLoc.GetLoc(args...)
}

func (this *LonglatLoc) GetLoc(args ...string) string {
	if len(args) != 3 {
		common.Error("wrong func argument")
		return ""
	}
	choice, long, lat := args[2], args[0], args[1]
	var (
		x, y     float64
		x_n, y_n int
		code     string
		loc      string
		ok       bool
		xy       string
	)

	x, _ = strconv.ParseFloat(long, 10)
	y, _ = strconv.ParseFloat(lat, 10)

	x_n = int(x*1000-Start_x) / Step_x
	y_n = int(y*1000-Start_y) / Step_y
	xy = fmt.Sprintf("%v,%v", x_n, y_n)

	this.RLock()
	defer this.RUnlock()

	if code, ok = this.curLLMaps.codeMap[xy]; !ok {
		return Unknwon
	}

	if strings.HasPrefix(code, "110") ||
		strings.HasPrefix(code, "120") ||
		strings.HasPrefix(code, "310") ||
		strings.HasPrefix(code, "510") {
		code = fmt.Sprintf("%s000", code[:3])
	} else {
		code = fmt.Sprintf("%s00", code[:4])
	}

	switch choice {
	case City:
		if loc, ok = this.curLLMaps.cityMap[code]; ok {
			return loc
		}

	case Province:
		if loc, ok = this.curLLMaps.provinceMap[code]; ok {
			return loc
		}
	default:
	}
	return Unknwon
}

func InitLoc() {
	if GlobalLLLoc == nil {
		GlobalLLLoc = NewLongLatLoc()
	}
}

// 简历longlat 2 code 的映射建立
func buildCodeMap(filename string) (map[string]string, error) {
	mm := make(map[string]string)
	var (
		line     string
		x, y     float64
		x_n, y_n int
		xy       string
	)
	fobj, err := readFile(filename)
	if err != nil {
		return mm, err
	}
	for _, line = range fobj {
		parts := strings.Split(line, "\t")
		latlong, code := parts[0], parts[1]
		sparts := strings.Split(latlong, ",")

		x, _ = strconv.ParseFloat(sparts[0], 10)
		y, _ = strconv.ParseFloat(sparts[1], 10)

		x_n = int(x*1000-Start_x) / Step_x
		y_n = int(y*1000-Start_y) / Step_y

		xy = fmt.Sprintf("%v,%v", x_n, y_n)
		mm[xy] = code
	}
	return mm, nil
}

// 建立code to city映射
func buildCity(filename string) (map[string]string, error) {
	mm := make(map[string]string)
	var (
		line string
	)
	fobj, err := readFile(filename)
	if err != nil {
		return mm, err
	}
	for _, line = range fobj {
		parts := strings.Split(line, "\t")
		code, city := parts[0], parts[1]
		mm[code] = city
	}
	return mm, nil
}

// 建立code to province映射
func buildProvince(filename string) (map[string]string, error) {
	mm := make(map[string]string)
	var (
		line string
	)
	fobj, err := readFile(filename)
	if err != nil {
		return mm, err
	}
	for _, line = range fobj {
		parts := strings.Split(line, "\t")
		code, prov := parts[0], parts[1]

		if strings.HasPrefix(code, "110") ||
			strings.HasPrefix(code, "120") ||
			strings.HasPrefix(code, "310") ||
			strings.HasPrefix(code, "510") {
			code = fmt.Sprintf("%s000", code[:3])
		} else {
			code = fmt.Sprintf("%s00", code[:4])
		}
		mm[code] = prov
	}
	return mm, nil
}

func readFile(filename string) (content []string, err error) {
	var reader *bufio.Reader

	content = make([]string, 0)

	f, err := os.Open(filename)
	if err != nil {
		return content, err
	}
	defer f.Close()
	reader = bufio.NewReader(f)

	for {
		buf, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		buf = strings.TrimSpace(buf)
		content = append(content, buf)
	}
	return content, nil
}
