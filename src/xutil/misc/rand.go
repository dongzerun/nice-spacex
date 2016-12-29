package misc

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(n int) int {
	return rand.Intn(n)
}

func RandomInt64(n int64) int64 {
	return rand.Int63n(n)
}
