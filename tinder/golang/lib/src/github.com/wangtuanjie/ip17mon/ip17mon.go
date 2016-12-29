package ip17mon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
)

const Null = "N/A"

var (
	ErrInvalidIp = errors.New("invalid ip format")
	std          *Locator
)

// Init defaut locator with dataFile
func Init(dataFile string) (err error) {
	if std != nil {
		return
	}
	std, err = NewLocator(dataFile)
	return
}

// Init defaut locator with data
func InitWithData(data []byte) {
	if std != nil {
		return
	}
	std = NewLocatorWithData(data)
	return
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func Find(ipstr string) (*LocationInfo, error) {
	return std.Find(ipstr)
}

// Find locationInfo by uint32
func FindByUint(ip uint32) *LocationInfo {
	return std.FindByUint(ip)
}

//-----------------------------------------------------------------------------

// New locator with dataFile
func NewLocator(dataFile string) (loc *Locator, err error) {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return
	}
	loc = NewLocatorWithData(data)
	return
}

// New locator with data
func NewLocatorWithData(data []byte) (loc *Locator) {
	loc = new(Locator)
	loc.init(data)
	return
}

// 地址匹配结构体
type Locator struct {
	textData   []byte   // 存放的文本信息，国家省市ISP等最终展现给用户可用的信息
	indexData1 []uint32 // 索引信息
	indexData2 []int
	indexData3 []int
	index      []int
}

// 地址信息
type LocationInfo struct {
	Country string // 国家
	Region  string // 地区
	City    string // 市
	Isp     string // 运营商
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func (loc *Locator) Find(ipstr string) (info *LocationInfo, err error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		err = ErrInvalidIp
		return
	}
	info = loc.FindByUint(binary.BigEndian.Uint32([]byte(ip.To4())))
	return
}

// Find locationInfo by uint32
func (loc *Locator) FindByUint(ip uint32) (info *LocationInfo) {
	end := len(loc.indexData1) - 1
	if ip>>24 != 0xff {
		// 找到某一个A段在data中的末尾位置
		// 比如，213.12.53.33 那就找到214在data中的位置
		// 就是213段的末尾位置
		end = loc.index[(ip>>24)+1]
	}
	idx := loc.findIndexOffset(ip, loc.index[ip>>24], end)
	off := loc.indexData2[idx]
	return newLocationInfo(loc.textData[off : off+loc.indexData3[idx]])
}

// binary search
func (loc *Locator) findIndexOffset(ip uint32, start, end int) int {
	for start < end {
		mid := (start + end) / 2
		if ip > loc.indexData1[mid] {
			start = mid + 1
		} else {
			end = mid
		}
	}

	if loc.indexData1[end] >= ip {
		return end
	}

	return start
}

// 正常给人看的文本格式是这个
// 000.000.000.000 000.255.255.255 保留地址        保留地址        *       *       *       *
// 001.000.000.000 001.000.000.255 APNIC   APNIC   *       *       *       *
// 001.000.001.000 001.000.003.255 中国    福建    *       *       电信    *
// 001.000.004.000 001.000.006.255 澳大利亚        维多利亚州      墨尔本  *       *       *
// 001.000.007.000 001.000.007.255 澳大利亚        维多利亚州      墨尔本  *       *       *
// 001.000.008.000 001.000.015.255 中国    广东    *       *       电信    *
// 001.000.016.000 001.000.031.255 日本    日本    *       *       *       *
// 001.000.032.000 001.000.063.255 中国    广东    *       *       电信    *
// 001.000.064.000 001.000.127.255 日本    日本    *       *       *       *
// 001.000.128.000 001.000.255.255 泰国    泰国    *       *       *       *

// 我要反解这玩意
func (loc *Locator) init(data []byte) {
	// 二进制的开头 4字节是文本偏移位置
	textoff := int(binary.BigEndian.Uint32(data[:4]))

	// 文本偏移量前的 1024 字节也属于 textData ???
	loc.textData = data[textoff-1024:]

	// index 为啥是256个呢？？ 因为ip地址是 000~255
	loc.index = make([]int, 256)
	for i := 0; i < 256; i++ {
		off := 4 + i*4 // ip地址是4个，所以是  4 + i*4
		// 也就是说存储000，001，002，003 ~ 255 开头的ip在 data文件中的位置
		loc.index[i] = int(binary.LittleEndian.Uint32(data[off : off+4]))
	}

	nidx := (textoff - 4 - 1024 - 1024) / 8

	loc.indexData1 = make([]uint32, nidx)
	loc.indexData2 = make([]int, nidx)
	loc.indexData3 = make([]int, nidx)

	for i := 0; i < nidx; i++ {
		off := 4 + 1024 + i*8
		loc.indexData1[i] = binary.BigEndian.Uint32(data[off : off+4])
		loc.indexData2[i] = int(uint32(data[off+4]) | uint32(data[off+5])<<8 | uint32(data[off+6])<<16)
		loc.indexData3[i] = int(data[off+7])
	}
	return
}

func newLocationInfo(str []byte) *LocationInfo {

	var info *LocationInfo

	fields := bytes.Split(str, []byte("\t"))
	switch len(fields) {
	case 4:
		// free version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
		}
	case 5:
		// pay version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
			Isp:     string(fields[4]),
		}
	default:
		panic("unexpected ip info:" + string(str))
	}

	if len(info.Country) == 0 {
		info.Country = Null
	}
	if len(info.Region) == 0 {
		info.Region = Null
	}
	if len(info.City) == 0 {
		info.City = Null
	}
	if len(info.Isp) == 0 {
		info.Isp = Null
	}
	return info
}
