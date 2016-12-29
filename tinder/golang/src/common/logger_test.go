package common_test

import (
	. "common"
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := NewDefaultLogger()
	logger.Info(fmt.Sprintf("xxxxxxxxxxx"))
}
