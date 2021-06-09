package stack

import (
	"fmt"
	"testing"
	"time"
)

func Test_SafeGo(t *testing.T) {

	testFunc := func() error {

		for i := 0; i < 3; i++ {
			fmt.Println("---------------")
			var dd []int
			dd[3] = 0
		}

		return nil
	}

	SafeGo(testFunc)

	time.Sleep(5 * time.Second)
}
