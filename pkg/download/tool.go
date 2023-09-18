package download

import (
	"os"
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
			GetGithubBaseUrl()+name+"/"+name+".7z",
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
		Version: "1.5.5",
		mac:     "hL0EMVepIyplxO4c8ZbESm6eGBs8IRMybyk81b76nLk6wHM4dXN9mi7CPmTAMa6gw06ki6Vr4w6vI69+HvIKGg==",
		linux: map[string]string{
			"x64": "01M9lAhvtX50Lb0CNZ4mY3ajGTVvKwlbDNLjE/e93lg9AfYFDNG5C9twCKbvvrXjatDCT6w3eCCFw0tw5221RA==",
		},
		win: map[string]string{
			"ia32": "jddFtdnYsgXmm9qozFHYqIry8fPlr61ytnKDXV+d7w/HIe4E6kCBZholADqIrGFgcCmblhY4Nh/t8oBTLE7eYQ==",
			"x64":  "Cg/7RInWfRhfibx4TJ1SMgw5LMeFQp6lH0GA9CP1/EhlE+RomYc1yKJhwDMnO31s0841feZbqdcHTPhQTQyfDg==",
		},
	}, osName)
}

func DownloadWinCodeSign() (string, error) {
	//noinspection SpellCheckingInspection
	return downloadFromGithub("winCodeSign", "2.6.0", "6LQI2d9BPC3Xs0ZoTQe1o3tPiA28c7+PY69Q9i/pD8lY45psMtHuLwv3vRckiVr3Zx1cbNyLlBR8STwCdcHwtA==")
}

func downloadFromGithub(name string, version string, checksum string) (string, error) {
	id := name + "-" + version
	return DownloadArtifact(id, GetGithubBaseUrl()+id+"/"+id+".7z", checksum)
}

func GetGithubBaseUrl() string {
	v := os.Getenv("NPM_CONFIG_ELECTRON_BUILDER_BINARIES_MIRROR")
	if len(v) == 0 {
		v = os.Getenv("npm_config_electron_builder_binaries_mirror")
	}
	if len(v) == 0 {
		v = os.Getenv("npm_package_config_electron_builder_binaries_mirror")
	}
	if len(v) == 0 {
		v = os.Getenv("ELECTRON_BUILDER_BINARIES_MIRROR")
	}
	if len(v) == 0 {
		v = "https://github.com/electron-userland/electron-builder-binaries/releases/download/"
	}
	return v
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
