package config

import (
	"fmt"
	"testing"
	. "xutil/config"
)

func TestConfig(t *testing.T) {
	NewTomlConfig("example.toml")
	fmt.Println(conf)
}
