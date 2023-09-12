package node_modules

import (
	"path"
	"runtime"
)

func Dirname() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
