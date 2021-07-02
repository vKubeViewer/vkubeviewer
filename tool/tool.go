package tool

import (
	"fmt"
)

// this brilliant code is from https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
// change it to 2 digts float
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func ArrayEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for i, v := range first {
		if second[i] != v {
			return false
		}
	}
	return true
}
