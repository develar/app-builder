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

func nodeModuleExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	packageJsonPath := filepath.Join(path, "package.json")
	_, err = os.Stat(packageJsonPath)
	return err == nil
}

func FindParentNodeModuleWithFile(cwd string, file string) string {
	if nodeModuleExists(path.Join(cwd, file)) {
		return cwd
	}

	parent := filepath.Dir(cwd)
	if parent == cwd {
		return ""
	}
	return FindParentNodeModuleWithFile(parent, file)
}
