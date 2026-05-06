// Package waiter polls a process by PID and reports when it disappears.
package waiter

import (
	"fmt"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

// Waiter polls the OS process table for a given PID.
type Waiter struct {
	Logger *logrus.Logger
}

// Wait returns a channel that is closed once the process with the given PID
// is no longer present. It polls every `timeout` milliseconds.
func (w *Waiter) Wait(pid int, timeout int64) (<-chan struct{}, error) {
	pc, err := ps.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("find process %d: %w", pid, err)
	}
	if pc == nil {
		return nil, fmt.Errorf("process with PID %d not found", pid)
	}
	w.debugf("watching pid=%d (%s) every %dms", pid, pc.Executable(), timeout)

	out := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(out)
		ticks := 0
		for {
			ticks++
			proc, perr := ps.FindProcess(pid)
			if perr != nil {
				w.debugf("pid=%d poll #%d: error: %v", pid, ticks, perr)
			} else if proc == nil {
				w.infof("pid=%d disappeared after %s (%d polls)", pid, time.Since(start).Round(time.Millisecond), ticks)
				return
			} else {
				w.debugf("pid=%d poll #%d: still alive (%s)", pid, ticks, proc.Executable())
			}
			time.Sleep(time.Duration(timeout) * time.Millisecond)
		}
	}()
	return out, nil
}

func (w *Waiter) debugf(format string, args ...interface{}) {
	if w.Logger == nil {
		return
	}
	w.Logger.Debugf(format, args...)
}

func (w *Waiter) infof(format string, args ...interface{}) {
	if w.Logger == nil {
		return
	}
	w.Logger.Infof(format, args...)
}
