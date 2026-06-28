//go:build darwin

package waiter

import (
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// processStartTime returns the process start time (microseconds since the
// epoch) as a token that is stable for the lifetime of the process and changes
// if the PID is reused. It queries kern.proc.pid via sysctl and reads the
// timeval at the head of the returned kinfo_proc: the struct begins with
// extern_proc, whose first union member is `struct timeval p_starttime`, which
// the kernel fills with the process start time. Returns errProcessGone when the
// process no longer exists.
func processStartTime(pid int) (uint64, error) {
	buf, err := unix.SysctlRaw("kern.proc.pid", pid)
	if err != nil {
		// macOS reports a vanished PID via ESRCH/ENOENT on some paths.
		if errors.Is(err, unix.ESRCH) || errors.Is(err, unix.ENOENT) {
			return 0, errProcessGone
		}
		return 0, fmt.Errorf("sysctl kern.proc.pid.%d: %w", pid, err)
	}
	// An empty buffer means there is no such process.
	if len(buf) == 0 {
		return 0, errProcessGone
	}
	// struct timeval { int64 tv_sec; int32 tv_usec } occupies the first
	// 12 meaningful bytes of the kinfo_proc buffer (followed by padding).
	if len(buf) < 12 {
		return 0, fmt.Errorf("short kinfo_proc for pid %d: %d bytes", pid, len(buf))
	}
	sec := binary.LittleEndian.Uint64(buf[0:8])
	usec := binary.LittleEndian.Uint32(buf[8:12])
	return sec*1_000_000 + uint64(usec), nil
}
