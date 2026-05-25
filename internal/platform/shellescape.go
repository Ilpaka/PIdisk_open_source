package platform

import "strings"

// ShellEscape quotes s for use as a single POSIX-shell argument.
// We use it only for the `df` fallback path; everything else uses native SFTP calls.
func ShellEscape(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
