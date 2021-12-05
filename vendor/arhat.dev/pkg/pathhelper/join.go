package pathhelper

import (
	"path/filepath"
	"strings"
)

func JoinUnixPath(elem ...string) string {
	// If there's a bug here, fix the logic in ./path_plan9.go too.
	for i, e := range elem {
		if e != "" {
			return filepath.Clean(strings.Join(elem[i:], string(UnixPathSeparator)))
		}
	}
	return ""
}

func JoinWindowsPath(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return CleanWindowsPath(joinNonEmpty(elem[i:]))
		}
	}
	return ""
}

// joinNonEmpty is like join, but it assumes that the first element is non-empty.
func joinNonEmpty(elem []string) string {
	if len(elem[0]) == 2 && elem[0][1] == ':' {
		// First element is drive letter without terminating slash.
		// Keep path relative to current directory on that drive.
		// Skip empty elements.
		i := 1
		for ; i < len(elem); i++ {
			if elem[i] != "" {
				break
			}
		}
		return CleanWindowsPath(elem[0] + strings.Join(elem[i:], string(WindowsPathSeparator)))
	}
	// The following logic prevents Join from inadvertently creating a
	// UNC path on Windows. Unless the first element is a UNC path, Join
	// shouldn't create a UNC path. See golang.org/issue/9167.
	p := CleanWindowsPath(strings.Join(elem, string(WindowsPathSeparator)))
	if !isUNC(p) {
		return p
	}
	// p == UNC only allowed when the first element is a UNC path.
	head := CleanWindowsPath(elem[0])
	if isUNC(head) {
		return p
	}
	// head + tail == UNC, but joining two non-UNC paths should not result
	// in a UNC path. Undo creation of UNC path.
	tail := CleanWindowsPath(strings.Join(elem[1:], string(WindowsPathSeparator)))
	if head[len(head)-1] == WindowsPathSeparator {
		return head + tail
	}
	return head + string(WindowsPathSeparator) + tail
}
