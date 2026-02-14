//go:build linux || darwin || freebsd || netbsd || openbsd

package sandbox

import (
	"os"
	"runtime"
	"syscall"
)

func getProcessMaxRSSKB(state *os.ProcessState) int {
	if state == nil {
		return 0
	}

	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok || usage == nil {
		return 0
	}

	maxRSS := int(usage.Maxrss)
	if maxRSS <= 0 {
		return 0
	}

	// macOS 的 Maxrss 单位是 bytes，其它 Unix 常见为 KB。
	if runtime.GOOS == "darwin" {
		return maxRSS / 1024
	}

	return maxRSS
}
