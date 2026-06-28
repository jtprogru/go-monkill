//go:build linux

package waiter

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// processStartTime returns the kernel start time of the process as a token that
// is stable for the lifetime of the process and changes if the PID is reused.
// It reads field 22 (starttime, in clock ticks since boot) from
// /proc/<pid>/stat. Returns errProcessGone when the process no longer exists.
func processStartTime(pid int) (uint64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, errProcessGone
		}
		return 0, err
	}

	// The comm field (2) is wrapped in parentheses and may itself contain
	// spaces or ')', so anchor parsing on the final ')'.
	s := string(data)
	rparen := strings.LastIndexByte(s, ')')
	if rparen < 0 || rparen+2 > len(s) {
		return 0, fmt.Errorf("malformed /proc/%d/stat", pid)
	}

	// Fields after comm start at field 3 (state) = index 0, so starttime
	// (field 22) sits at index 19.
	const starttimeIdx = 19
	fields := strings.Fields(s[rparen+2:])
	if len(fields) <= starttimeIdx {
		return 0, fmt.Errorf("unexpected /proc/%d/stat format: %d fields", pid, len(fields))
	}

	st, err := strconv.ParseUint(fields[starttimeIdx], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse starttime for pid %d: %w", pid, err)
	}
	return st, nil
}
