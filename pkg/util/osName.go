package util

import (
	"runtime"
)

type OsName int

const (
	MAC OsName = iota
	LINUX
	WINDOWS
)

func (t OsName) String() string {
	switch t {
	case MAC:
		return "mac"
	case WINDOWS:
		return "windows"
	default:
		return "linux"
	}
}

//noinspection GoExportedFuncWithUnexportedType
func GetCurrentOs() OsName {
	return ToOsName(runtime.GOOS)
}

//noinspection GoExportedFuncWithUnexportedType
func ToOsName(name string) OsName {
	switch name {
	case "windows", "win32", "win":
		return WINDOWS
	case "darwin", "mac", "macOS", "macOs":
		return MAC
	default:
		return LINUX
	}
}
