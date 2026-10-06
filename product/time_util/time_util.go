// Package timeutil holds internal duration helpers shared by product features.
// It has no customer-visible settings and no documentation mapping.
package timeutil

import "fmt"

// FormatSeconds renders a number of seconds as "1m30s" style text.
func FormatSeconds(total int) string {
	if total < 60 {
		return fmt.Sprintf("%ds", total)
	}
	return fmt.Sprintf("%dm%ds", total/60, total%60)
}
