//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd

package sandbox

import "os"

func getProcessMaxRSSKB(_ *os.ProcessState) int {
	return 0
}
