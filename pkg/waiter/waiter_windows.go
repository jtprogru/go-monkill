package waiter

import (
	"os"
)

type Waiter struct{}

func (w Waiter) Wait(pid int) (<-chan struct{}, error) {
	_, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	out := make(chan struct{})
	go func() {
		for {
			proc, _ := os.FindProcess(pid)
			if proc == nil {
				out <- struct{}{}
			}
		}
	}()
	return out, nil
}
