package download

import (
	"path/filepath"
	"runtime"

	"github.com/develar/errors"
)

func DownloadFpm() (string, error) {
	currentOs := GetCurrentOs()
	if currentOs == LINUX {
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
			"https://github.com/electron-userland/electron-builder-binaries/releases/download/" + name + "/" + name + ".7z",
			checksum,
		)
	} else {
		//noinspection SpellCheckingInspection
		return DownloadArtifact(
			"fpm-1.9.3-20150715-2.2.2-mac",
			"https://github.com/electron-userland/electron-builder-binaries/releases/download/fpm-1.9.3-20150715-2.2.2-mac/fpm-1.9.3-20150715-2.2.2-mac.7z",
			"oXfq+0H2SbdrbMik07mYloAZ8uHrmf6IJk+Q3P1kwywuZnKTXSaaeZUJNlWoVpRDWNu537YxxpBQWuTcF+6xfw==",
		)
	}
}

func DownloadZstd(osName osName) (string, error) {
	//noinspection SpellCheckingInspection
	return DownloadTool(ToolDescriptor{
		name: "zstd",
		version: "1.3.7",
		mac: "hTM6htzzi9ALEBdl2GBiH5NtmttBYEQiCugJ/u6CqsBwydPm39vgigSlUX1QMYGKu5jjG18vGKvjFJuhMQkOSw==",
		linux: map[string]string{
			"x64": "oSyW9a2YzLZ4xC8y7WyzivZcrpuc+NqN1/IESj7nqjUL9905N5ZfC50GbS6mSmu2nlVq5VNafFGgf/RUJf/pkA==",
		},
		win: map[string]string{
			"ia32": "EcogWukKij8MqHr0pEqFw8Wb3EHLeLgXyj+AbokJ5uejf8y38GyR4nbsSLlbstgkG0SbaJYs7JvyPFgaT2Hc0Q==",
			"x64": "NlN0f1Usvwdu4WRjK7KXfNPUOGv7uKJcwyGTX5fGL6zEuPJtCQ8cpd0S/Zgx1yxo0YYOFYm/z1xyy8eaoTxKog==",
		},
	}, osName)
}

func DownloadTool(descriptor ToolDescriptor, osName osName) (string, error) {
	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "armv7"
	} else if arch == "arm64" {
		arch = "armv8"
	} else if arch == "amd64" {
		arch = "x64"
	}

	var checksum string
	var archQualifier string
	var osQualifier string
	if osName == MAC {
		checksum = descriptor.mac
		archQualifier = ""
		osQualifier = "mac"
	} else {
		archQualifier = "-" + arch
		if osName == WINDOWS {
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
		tagPrefix = descriptor.name + "-"
	} else {
		tagPrefix = "v"
	}

	osAndArch := osQualifier + archQualifier
	return DownloadArtifact(
		descriptor.name+"-"+descriptor.version+"-"+osAndArch /* ability to use cache dir on any platform (e.g. keep cache under project) */,
		"https://github.com/"+repository+"/releases/download/"+tagPrefix+descriptor.version+"/"+descriptor.name+"-v"+descriptor.version+"-"+osAndArch+".7z",
		checksum,
	)
}

type ToolDescriptor struct {
	name string
	version string

	repository string

	mac string
	linux map[string]string
	win map[string]string
}

func GetZstd() (string, error) {
	dir, err := DownloadZstd(GetCurrentOs())
	if err != nil {
		return "", err
	}

	executableName := "zstd"
	if GetCurrentOs() == WINDOWS {
		executableName += ".exe"
	}

	return filepath.Join(dir, executableName), nil
}