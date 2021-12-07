package waiter

import "github.com/mitchellh/go-ps"

type Waiter struct{}

func (w Waiter) Wait(pid int) (<-chan struct{}, error) {
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
		}
	}()
	return out, nil
}
