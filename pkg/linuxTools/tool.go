package linuxTools

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
)

func GetAppImageToolDir() (string, error) {
	dirName := "appimage-9.1.0"
	//noinspection SpellCheckingInspection
	result, err := download.DownloadArtifact("",
		"https://github.com/electron-userland/electron-builder-binaries/releases/download/"+dirName+"/"+dirName+".7z",
		"/9ipJexioCIFK+aQ/LktAHEieFFWxwkikxXZZlKXzm3fY5tFs+xUKv2m4OymD6ITRGiA4zzAKmlWyhVOjCxXuw==")
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result, nil
}

func GetAppImageToolBin(toolDir string) string {
	if util.GetCurrentOs() == util.MAC {
		return filepath.Join(toolDir, "darwin")
	} else {
		return filepath.Join(toolDir, "linux-"+goArchToNodeArch(runtime.GOARCH))
	}
}

func GetLinuxTool(name string) (string, error) {
	toolDir, err := GetAppImageToolDir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Join(GetAppImageToolBin(toolDir), name), nil
}

func GetMksquashfs() (string, error) {
	var err error

	result := "mksquashfs"
	if !util.IsEnvTrue("USE_SYSTEM_MKSQUASHFS") {
		result = os.Getenv("MKSQUASHFS_PATH")
		if len(result) == 0 {
			result, err = GetLinuxTool("mksquashfs")
			if err != nil {
				return "", errors.WithStack(err)
			}
		}
	}

	return result, nil
}

func goArchToNodeArch(arch string) string {
	switch arch {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	default:
		return arch
	}
}

func ReadDirContentTo(dir string, paths []string, filter func(string) bool) ([]string, error) {
	content, err := fsutil.ReadDirContent(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, value := range content {
		if filter == nil || filter(value) {
			paths = append(paths, filepath.Join(dir, value))
		}
	}
	return paths, nil
}
