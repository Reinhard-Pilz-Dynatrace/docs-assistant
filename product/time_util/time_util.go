// Package timeutil holds internal duration helpers shared by product features.
// It has no customer-visible settings and no documentation mapping.
package timeutil

import "fmt"

// FormatSeconds renders a number of seconds as "1m30s" style text.
func FormatSeconds(total int) string {
	minutes, seconds := total/60, total%60
	if minutes == 0 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}
