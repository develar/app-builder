package proton_native

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"go.uber.org/zap"
)

type ProtonNativeOptions struct {
	nodeJsVersion   string
	LaunchUiVersion string

	stageDir       string
	executableName string

	platform util.OsName
	arch     string

	isUseLaunchUi bool
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("proton-native", "Package Proton Native")

	version := command.Flag("node-version", "").Required().String()
	isUseLaunchUi := command.Flag("use-launch-ui", "").Default("false").Bool()
	platform := command.Flag("platform", "").Required().Enum("darwin", "linux", "win32")
	arch := command.Flag("arch", "").Default("x64").Enum("x64", "ia32")

	stageDir := command.Flag("stage", "Stage dir").Required().String()
	executableName := command.Flag("executable", "The application executable name").String()

	command.Action(func(context *kingpin.ParseContext) error {
		err := pack(ProtonNativeOptions{
			nodeJsVersion: *version,

			platform: util.ToOsName(*platform),
			arch:     *arch,

			stageDir:       *stageDir,
			executableName: *executableName,

			isUseLaunchUi: *isUseLaunchUi,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func pack(options ProtonNativeOptions) error {
	stageDir := options.stageDir
	if !options.isUseLaunchUi {
		nodeDir, err := downloadNodeJs(options.nodeJsVersion, options.arch, options.platform)
		if err != nil {
			return errors.WithStack(err)
		}
		executableName := toNodeJsExecutableName(options.platform)
		err = fs.CopyFileAndRestoreNormalPermissions(filepath.Join(nodeDir, executableName), filepath.Join(stageDir, executableName), 0755)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	if options.executableName == "" {
		return util.NewMessageError("executableName is empty", "EXECUTABLE_NAME_EMPTY")
	}

	unpackedLaunchUi, err := downloadLaunchUi(getLaunchUiVersion(options), options.platform, options.arch)
	if err != nil {
		return errors.WithStack(err)
	}

	var fileCopier fs.FileCopier
	fileCopier.IsUseHardLinks = false
	err = fileCopier.CopyDirOrFile(unpackedLaunchUi, stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	skeletonExecutable := "launchui"
	if options.platform == util.WINDOWS {
		skeletonExecutable += ".exe"
	}

	err = os.Rename(filepath.Join(stageDir, skeletonExecutable), filepath.Join(stageDir, options.executableName))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func getLaunchUiVersion(options ProtonNativeOptions) string {
	if options.LaunchUiVersion != "" {
		return options.LaunchUiVersion
	} else {
		// todo grab corresponding for NodeJS version from GitHub
		return "0.1.4-10.13.0"
	}
}

func downloadLaunchUi(version string, platform util.OsName, arch string) (string, error) {
	checksum := ""
	if version == "0.1.4-10.13.0" {
		if arch == "ia32" {
			//noinspection SpellCheckingInspection
			checksum = "Ha4WpmVqFR7KmAOvYs/n8PLXmwy50EWGV6s4qb5A5Ib+FReYvpWT/xZ23I6Xh9WyzS7+CpmrEGEVDOkmtMKK/w=="
		} else {
			if platform == util.WINDOWS {
				//noinspection SpellCheckingInspection
				checksum = "sBzi/o4sHajG5/TDZOzHcZ4V34SCekb6bm71fvzy+UbsPGQbOtKNFh2dEIgYgK9vxN+LVHbx1i5Xq7FbcaLnEQ=="
			} else {
				//noinspection SpellCheckingInspection
				checksum = "Ip8zGEW3jBs2aRCTYqB44bFT5LkOYS2JghSEPvWar+0DEwibwhSVSiF0Uz+ONt0ug8jAtOuPQn43ZxYT6mKhbQ=="
			}
		}
	}

	name := "v" + version + "-" + toLaunchUiDownloadPlatform(platform) + "-" + arch
	return download.DownloadArtifact(
		"launchui-"+name,
		"https://github.com/develar/launchui/releases/download/v"+version+"/launchui-"+name+".7z",
		checksum)
}

func toNodeJsDownloadPlatform(os util.OsName) string {
	switch os {
	case util.MAC:
		return "darwin"

	case util.WINDOWS:
		return "win"

	default:
		return "linux"
	}
}

func toLaunchUiDownloadPlatform(os util.OsName) string {
	switch os {
	case util.MAC:
		return "mac"

	case util.WINDOWS:
		return "win32"

	default:
		return "linux"
	}
}

func toNodeJsExecutableName(os util.OsName) string {
	if os == util.WINDOWS {
		return "node.exe"
	} else {
		return "node"
	}
}

func downloadNodeJs(version string, arch string, platform util.OsName) (string, error) {
	var format string
	if platform == util.WINDOWS {
		format = "7z"
	} else {
		format = "tar.xz"
	}

	cacheDir, err := download.GetCacheDirectoryForArtifactCustom("node")
	if err != nil {
		return "", errors.WithStack(err)
	}

	dirPath := filepath.Join(cacheDir, version+"-"+toNodeJsDownloadPlatform(platform)+"-"+arch)

	logger := log.LOG.With(zap.String("path", dirPath))

	isFound, err := download.CheckCache(dirPath, cacheDir, logger)
	if isFound {
		return dirPath, nil
	}
	if err != nil {
		return "", errors.WithStack(err)
	}

	url := getNodeJsDownloadUrl(version, platform, arch, format)

	// 7z cannot be extracted from the input stream, temp file is required
	tempUnpackDir, err := util.TempDir(cacheDir, "")
	if err != nil {
		return "", errors.WithStack(err)
	}

	archiveName := tempUnpackDir + "." + format

	err = download.NewDownloader().Download(url, archiveName, "")
	if err != nil {
		return "", errors.WithStack(err)
	}

	if format == "tar.xz" {
		err = unpackTarXzNodeJs(archiveName, tempUnpackDir)
		if err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		command := exec.Command(util.Get7zPath(), "e", "-bd", archiveName, "-o"+tempUnpackDir, "*/node.exe", "-r")
		command.Dir = cacheDir
		output, err := command.CombinedOutput()
		if err != nil {
			return "", errors.WithStack(err)
		}

		log.Debug(string(output))
	}

	download.RemoveArchiveFile(archiveName, tempUnpackDir, logger)
	download.RenameToFinalFile(tempUnpackDir, dirPath, logger)

	return dirPath, nil
}

func getNodeJsDownloadUrl(version string, platform util.OsName, arch string, format string) string {
	if arch == "ia32" {
		arch = "x86"
	}
	return "https://nodejs.org/dist/v" + version + "/node-v" + version + "-" + toNodeJsDownloadPlatform(platform) + "-" + arch + "." + format
}

func unpackTarXzNodeJs(archiveName string, unpackDir string) error {
	decompressCommand := exec.Command(util.Get7zPath(), "e", "-bd", "-txz", archiveName, "-so")

	//noinspection SpellCheckingInspection
	unTarCommand := exec.Command(util.Get7zPath(), "e", "-bd", "-ttar", "-o"+unpackDir, "*/bin/node", "-r", "-si")
	err := download.RunExtractCommands(decompressCommand, unTarCommand)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Chmod(filepath.Join(unpackDir, "node"), 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
