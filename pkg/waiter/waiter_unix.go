package waiter

import (
	"fmt"
	"os"
)

type Waiter struct{}

const procPath = "/proc/%d"

func (w Waiter) checkProcess(pid int) bool {
	_, err := os.Stat(fmt.Sprintf(procPath, pid))
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (w Waiter) Wait(pid int) (<-chan struct{}, error) {
	out := make(chan struct{})
	go func() {
		var ok bool
		for {
			ok = w.checkProcess(pid)
			if !ok {
				out <- struct{}{}
			}
		}
	}()
	return out, nil
}
