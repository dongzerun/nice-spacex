package xplugin

import (
	"strconv"
	"strings"
)

var (
	StrIntMap map[string]int
)

func init() {
	if StrIntMap == nil {
		StrIntMap = make(map[string]int)
		for i := 0; i < 10000; i++ {
			str := strconv.Itoa(i)
			StrIntMap[str] = i
		}
	}
}

func ParseStrToInt(s string) (int, error) {
	i, ok := StrIntMap[s]
	if ok {
		return i, nil
	}
	return strconv.Atoi(s)
}

func ParseIntToStr(i int) string {
	return strconv.Itoa(i)
}

func SimpleStrMatch(data []string, str []string) bool {
	for i, _ := range data {
		if simpleStrMatch(str, data[i]) {
			return true
		}
	}
	return false
}

func SimpleIntMatch(data []int, target int) bool {
	for i, _ := range data {
		if data[i] == target {
			return true
		}
	}
	return false
}

func SimpleInt64Match(data []int64, target int64) bool {
	for i, _ := range data {
		if data[i] == target {
			return true
		}
	}
	return false
}

func simpleStrMatch(data []string, str string) bool {
	for i, _ := range data {
		if data[i] == str {
			return true
		}
	}
	return false
}

func VersionRangeMatch(ver [][]int, target []int) bool {
	if len(ver) != 2 {
		//version必须两个元素，起始版本号和末尾版本号，是一个范围
		return false
	}

	match := true

	l := len(ver[0])
	if len(target) < l {
		l = len(target)
	}

	for p := 0; p < l; p++ {
		// 匹配版本号
		if target[p] < ver[0][p] || target[p] > ver[1][p] {
			match = false
			break
		}
	}

	return match
}

func VersionEnumMatch(ver [][]int, target []int) bool {

	for i := 0; i < len(ver); i++ {
		match := true
		l := len(ver[i])
		if len(target) < l {
			l = len(target)
		}

		for p := 0; p < l; p++ {
			// 匹配版本号
			if ver[i][p] != target[p] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func IntSlice(str string) ([]int, error) {
	s := strings.Split(str, ".")

	n := make([]int, len(s))
	for i, _ := range s {
		is, err := strconv.Atoi(s[i])
		if err != nil {
			return nil, err
		}
		n[i] = is
	}
	return n, nil
}
