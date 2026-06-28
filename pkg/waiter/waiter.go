// Package waiter polls a process by PID and reports when it disappears.
package waiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

// errProcessGone is returned by processStartTime (see waiter_linux.go /
// waiter_darwin.go) when the PID no longer maps to a live process.
var errProcessGone = errors.New("process no longer exists")

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

	// Capture the process start time so we can detect PID reuse: a PID is only
	// unique together with the moment its current owner started. If the watched
	// process exits and the kernel hands the same numeric PID to an unrelated
	// process, the start time changes and we treat the original as terminated
	// instead of latching onto the impostor.
	origStart, err := processStartTime(pid)
	if err != nil {
		return nil, fmt.Errorf("read start time for pid %d: %w", pid, err)
	}
	w.debugf("watching pid=%d (%s) start=%d every %dms", pid, pc.Executable(), origStart, timeout)

	out := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(out)
		ticker := time.NewTicker(time.Duration(timeout) * time.Millisecond)
		defer ticker.Stop()
		ticks := 0
		for {
			ticks++
			curStart, perr := processStartTime(pid)
			switch {
			case errors.Is(perr, errProcessGone):
				w.infof("pid=%d disappeared after %s (%d polls)", pid, time.Since(start).Round(time.Millisecond), ticks)
				return
			case perr != nil:
				w.debugf("pid=%d poll #%d: error: %v", pid, ticks, perr)
			case curStart != origStart:
				w.infof("pid=%d reused by a different process after %s (%d polls); treating original as exited",
					pid, time.Since(start).Round(time.Millisecond), ticks)
				return
			default:
				w.debugf("pid=%d poll #%d: still alive", pid, ticks)
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
