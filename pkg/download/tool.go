package download

import (
	"path/filepath"
	"runtime"

	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
)

func DownloadFpm() (string, error) {
	currentOs := util.GetCurrentOs()
	if currentOs == util.LINUX {
		var checksum string
		var archSuffix string
		if runtime.GOARCH == "amd64" {
			//noinspection SpellCheckingInspection
			checksum = "fcKdXPJSso3xFs5JyIJHG1TfHIRTGDP0xhSBGZl7pPZlz4/TJ4rD/q3wtO/uaBBYeX0qFFQAFjgu1uJ6HLHghA=="
			archSuffix = "-x86_64"
		} else {
			//noinspection SpellCheckingInspection
			checksum = "OnzvBdsHE5djcXcAT87rwbnZwS789ZAd2ehuIO42JWtBAHNzXKxV4o/24XFX5No4DJWGO2YSGQttW+zn7d/4rQ=="
			archSuffix = "-x86"
		}

		//noinspection SpellCheckingInspection
		name := "fpm-1.9.3-2.3.1-linux" + archSuffix
		return DownloadArtifact(
			name,
			"https://github.com/electron-userland/electron-builder-binaries/releases/download/"+name+"/"+name+".7z",
			checksum,
		)
	} else {
		//noinspection SpellCheckingInspection
		return downloadFromGithub("fpm", "1.9.3-20150715-2.2.2-mac", "oXfq+0H2SbdrbMik07mYloAZ8uHrmf6IJk+Q3P1kwywuZnKTXSaaeZUJNlWoVpRDWNu537YxxpBQWuTcF+6xfw==")
	}
}

func DownloadZstd(osName util.OsName) (string, error) {
	//noinspection SpellCheckingInspection
	return DownloadTool(ToolDescriptor{
		Name:    "zstd",
		Version: "1.4.0",
		mac:     "CY+P8Egn6V14gcWFHz3hKnpKAn7/18PxzotcMXnM3CREBiveygAojJxlRJ9PsAqLlGFHiAd1SGOv/IhLwrKvHQ==",
		linux: map[string]string{
			"x64": "T09y1u1YwOp21/AdH4ojfBna6Xx2/IVk70nJEuTULFBULY84/WHub/19/c1+P7jr3HlXOVpUkOmnKp3BCBR00g==",
		},
		win: map[string]string{
			"ia32": "2/NjLk7LTZUl2mPczgv92dWM2gZjhkPuPBmFxDZvOJC20JmvKsiJj+e5afUazhIiKAMdWS0PmthvoCYJWFbdJg==",
			"x64":  "Rj7thtr+5anAX2NKxbfvgohF4MEg6W1FVPahodfhby/BZU/7wHvq7GuHYZiaOguZB4ONJEBm/sI14ICaPYaNYg==",
		},
	}, osName)
}

func DownloadWinCodeSign() (string, error) {
	//noinspection SpellCheckingInspection
	return downloadFromGithub("winCodeSign", "2.5.0", "xoTtj+PXTqTs47tkB/DyLKkXMFAclwRS3JNPOaZu7ZBnvs/gbY33ZSi+x2CH0xB83LAm+u6ixlhxtcMTl1Vrjg==")
}

func downloadFromGithub(name string, version string, checksum string) (string, error) {
	id := name + "-" + version
	return DownloadArtifact(id, "https://github.com/electron-userland/electron-builder-binaries/releases/download/"+id+"/"+id+".7z", checksum)
}

func DownloadTool(descriptor ToolDescriptor, osName util.OsName) (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "arm":
		//noinspection SpellCheckingInspection
		arch = "armv7"
	case "arm64":
		//noinspection SpellCheckingInspection
		arch = "armv8"
	case "amd64":
		arch = "x64"
	}

	var checksum string
	var archQualifier string
	var osQualifier string
	if osName == util.MAC {
		checksum = descriptor.mac
		archQualifier = ""
		osQualifier = "mac"
	} else {
		archQualifier = "-" + arch
		if osName == util.WINDOWS {
			osQualifier = "win"
			checksum = descriptor.win[arch]
		} else {
			osQualifier = "linux"
			checksum = descriptor.linux[arch]
		}
	}

	if checksum == "" {
		return "", errors.Errorf("Checksum not specified for %s:%s", osName, arch)
	}

	repository := descriptor.repository
	if repository == "" {
		repository = "electron-userland/electron-builder-binaries"
	}

	var tagPrefix string
	if descriptor.repository == "" {
		tagPrefix = descriptor.Name + "-"
	} else {
		tagPrefix = "v"
	}

	osAndArch := osQualifier + archQualifier
	return DownloadArtifact(
		descriptor.Name+"-"+descriptor.Version+"-"+osAndArch, /* ability to use cache dir on any platform (e.g. keep cache under project) */
		"https://github.com/"+repository+"/releases/download/"+tagPrefix+descriptor.Version+"/"+descriptor.Name+"-v"+descriptor.Version+"-"+osAndArch+".7z",
		checksum,
	)
}

type ToolDescriptor struct {
	Name    string
	Version string

	repository string

	mac   string
	linux map[string]string
	win   map[string]string
}

func GetZstd() (string, error) {
	dir, err := DownloadZstd(util.GetCurrentOs())
	if err != nil {
		return "", err
	}

	executableName := "zstd"
	if util.GetCurrentOs() == util.WINDOWS {
		executableName += ".exe"
	}

	return filepath.Join(dir, executableName), nil
}
