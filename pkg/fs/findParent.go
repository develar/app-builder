package fs

import (
	"os"
	"path"
	"path/filepath"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FindParentWithFile(cwd string, file string) string {
	if pathExists(path.Join(cwd, file)) {
		return cwd
	}
	parent := filepath.Dir(cwd)
	if parent == cwd {
		return ""
	}
	return FindParentWithFile(parent, file)
}
