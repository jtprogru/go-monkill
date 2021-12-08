package waiter

import (
	"github.com/mitchellh/go-ps"
	"time"
)

// Waiter struct
type Waiter struct{}

// Wait find process with defined PID and wait for process will finish or be killed.
// Checking the liveliness of the process occurs with a timeout delay.
// The timeout is set in milliseconds.
func (w Waiter) Wait(pid int, timeout int64) (<-chan struct{}, error) {
	_, err := ps.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	out := make(chan struct{})
	go func() {
		for {
			if pc, _ := ps.FindProcess(pid); pc == nil {
				out <- struct{}{}
			}
			time.Sleep(time.Duration(timeout) / time.Second)
		}
	}()
	return out, nil
}
