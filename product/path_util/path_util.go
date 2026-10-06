// Package pathutil holds internal path helpers shared by product features.
// It has no customer-visible settings and no documentation mapping.
package pathutil

import "strings"

// NormalizeSlashes replaces backslashes with forward slashes.
func NormalizeSlashes(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
