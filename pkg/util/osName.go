package util

import "runtime"

type OsName int

const (
	MAC OsName = iota
	LINUX
	WINDOWS
)

//noinspection GoExportedFuncWithUnexportedType
func GetCurrentOs() OsName {
	return ToOsName(runtime.GOOS)
}

//noinspection GoExportedFuncWithUnexportedType
func ToOsName(name string) OsName {
	if name == "windows" || name == "win32" {
		return WINDOWS
	} else if name == "darwin" || name == "mac" || name == "macOs" {
		return MAC
	} else {
		return LINUX
	}
}