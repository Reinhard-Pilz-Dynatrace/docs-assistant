// Package textutil holds internal string helpers shared by product features.
// It has no customer-visible settings and no documentation mapping.
package textutil

import "strings"

// Truncate shortens text to at most limit bytes, appending an ellipsis marker.
func Truncate(text string, limit int) string {
	if limit < 0 || len(text) <= limit {
		return text
	}
	return strings.TrimRight(text[:limit], " ") + "..."
}
