package waiter

import "github.com/mitchellh/go-ps"

//Waiter - process waiter
type Waiter struct{}

//Wait - return nonbuffered chan. Write on chan when process with specified pid doned
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
