// Package waiter polls a process by PID and reports when it disappears.
package waiter

import (
	"context"
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
// is no longer present, or once ctx is canceled. It polls every `timeout`
// milliseconds.
func (w *Waiter) Wait(ctx context.Context, pid int, timeout int64) (<-chan struct{}, error) {
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
		ticker := time.NewTicker(time.Duration(timeout) * time.Millisecond)
		defer ticker.Stop()
		ticks := 0
		for {
			ticks++
			proc, perr := ps.FindProcess(pid)
			switch {
			case perr != nil:
				w.debugf("pid=%d poll #%d: error: %v", pid, ticks, perr)
			case proc == nil:
				w.infof("pid=%d disappeared after %s (%d polls)", pid, time.Since(start).Round(time.Millisecond), ticks)
				return
			default:
				w.debugf("pid=%d poll #%d: still alive (%s)", pid, ticks, proc.Executable())
			}
			select {
			case <-ctx.Done():
				w.infof("pid=%d watch canceled after %s (%d polls)", pid, time.Since(start).Round(time.Millisecond), ticks)
				return
			case <-ticker.C:
			}
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
