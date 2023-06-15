package fs

import (
	"path/filepath"
)

// Predicate is a function that takes a directory and returns a boolean.
type Predicate func(dir string) bool

// FindParent recursively searches for a parent directory that satisfies the predicate.
func FindParent(cwd string, predicate Predicate) string {
	if predicate(cwd) {
		return cwd
	}
	parent := filepath.Dir(cwd)
	if parent == cwd {
		return ""
	}
	return FindParent(parent, predicate)
}
