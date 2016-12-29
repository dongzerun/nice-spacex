package common_test

import (
	. "common"
	"fmt"
	"runtime"
	"testing"
)

func TestRmq(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := "rmq.cfg"
	mqpool, err := NewMQPool(config, "local_rmq")
	if err != nil {
		panic(err)
	}

	defer mqpool.Close()

	msgs, err := mqpool.SetEx()

	if err != nil {
		panic(err)
	}

	forever := make(chan bool)

	n := 0
	for {
		if n > mqpool.ConsumeParallel {
			break
		}
		go func() {
			for msg := range msgs {
				fmt.Printf("1--------%s\n", msg.Body)
				msg.Ack(false)
			}
		}()
		n++
	}

	<-forever
}
