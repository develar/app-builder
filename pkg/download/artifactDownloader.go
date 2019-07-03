package download

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
)

func ConfigureArtifactCommand(app *kingpin.Application) {
	command := app.Command("download-artifact", "Download, unpack and cache artifact from GitHub.")
	name := command.Flag("name", "The artifact name.").Short('n').Required().String()
	url := command.Flag("url", "The artifact URL.").Short('u').String()
	sha512 := command.Flag("sha512", "The expected sha512 of file.").String()

	command.Action(func(context *kingpin.ParseContext) error {
		dirPath, err := DownloadArtifact(*name, *url, *sha512)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = os.Stdout.Write([]byte(dirPath))
		return errors.WithStack(err)
	})
}

func GetCacheDirectoryForArtifact(dirName string) (string, error) {
	result, err := GetCacheDirectory("electron-builder", "ELECTRON_BUILDER_CACHE", true)
	if err != nil {
		return "", errors.WithStack(err)
	}

	hyphenIndex := strings.IndexRune(dirName, '-')
	if hyphenIndex > 0 {
		result = filepath.Join(result, dirName[0:hyphenIndex])
	} else {
		result = filepath.Join(result, dirName)
	}
	return result, nil
}

func GetCacheDirectoryForArtifactCustom(dirName string) (string, error) {
	result, err := GetCacheDirectory("electron-builder", "ELECTRON_BUILDER_CACHE", true)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Join(result, dirName), nil
}

// we cache in the global location - in the home dir, not in the node_modules/.cache (https://www.npmjs.com/package/find-cache-dir) because
// * don't need to find node_modules
// * don't pollute user project dir (important in case of 1-package.json project structure)
// * simplify/speed-up tests (don't download fpm for each test project)
func DownloadArtifact(dirName string, url string, checksum string) (string, error) {
	if len(url) == 0 {
		// if no url is provided download these artifacts from Github. Otherwise use the provided url to download the artifacts.
		switch dirName {
		case "fpm":
			return DownloadFpm()
		case "zstd":
			return DownloadZstd(util.GetCurrentOs())
		case "winCodeSign":
			return DownloadWinCodeSign()
		}
	}

	if len(dirName) == 0 {
		dirName = path.Base(url)
		// cannot simply find fist dot because file name can contains version like 9.1.0
		dirName = strings.TrimSuffix(dirName, ".7z")
		dirName = strings.TrimSuffix(dirName, ".tar")
	}

	cacheDir, err := GetCacheDirectoryForArtifact(dirName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	filePath := filepath.Join(cacheDir, dirName)
	logFields := log.LOG.With(zap.String("path", filePath))

	isFound, err := CheckCache(filePath, cacheDir, logFields)
	if isFound {
		return filePath, nil
	}
	if err != nil {
		return "", err
	}

	// 7z cannot be extracted from the input stream, temp file is required
	tempUnpackDir, err := util.TempDir(cacheDir, "")
	if err != nil {
		return "", err
	}

	archiveName := tempUnpackDir + ".7z"

	err = NewDownloader().Download(url, archiveName, checksum)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(url, ".tar.7z") {
		err = unpackTar7z(archiveName, tempUnpackDir)
		if err != nil {
			return "", err
		}
	} else {
		command := exec.Command(util.Get7zPath(), "x", "-bd", archiveName, "-o"+tempUnpackDir)
		command.Dir = cacheDir
		_, err := util.Execute(command)
		if err != nil {
			return "", err
		}
	}

	RemoveArchiveFile(archiveName, tempUnpackDir, logFields)
	RenameToFinalFile(tempUnpackDir, filePath, logFields)

	return filePath, nil
}

func RemoveArchiveFile(archiveName string, tempUnpackDir string, logger *zap.Logger) {
	err := os.Remove(archiveName)
	if err != nil {
		logger.Warn("cannot remove downloaded archive (another process downloaded faster?)", zap.String("tempUnpackDir", tempUnpackDir), zap.Error(err))
	}
}

func CheckCache(filePath string, cacheDir string, logger *zap.Logger) (bool, error) {
	dirStat, err := os.Stat(filePath)
	if err == nil && dirStat.IsDir() {
		logger.Debug("found existing")
		return true, nil
	}

	if err != nil && !os.IsNotExist(err) {
		return false, errors.WithMessage(err, "error during cache check for path "+filePath)
	}

	err = fsutil.EnsureDir(cacheDir)
	if err != nil {
		return false, err
	}

	return false, nil
}

func RenameToFinalFile(tempFile string, filePath string, logger *zap.Logger) {
	err := os.Rename(tempFile, filePath)
	if err != nil {
		logger.Warn("cannot move downloaded into final location (another process downloaded faster?)", zap.String("tempFile", tempFile), zap.Error(err))
	}
}

func unpackTar7z(archiveName string, unpackDir string) error {
	decompressCommand := exec.Command(util.Get7zPath(), "e", "-bd", "-t7z", archiveName, "-so")

	args := []string{"-x"}
	//noinspection SpellCheckingInspection
	if runtime.GOOS == "darwin" {
		// otherwise snap error review "unusual mode 'rwxr-xr-x' for symlink"
		args = append(args, "-p")
	}
	args = append(args, "-f", "-")

	//noinspection SpellCheckingInspection
	unTarCommand := exec.Command("tar", args...)
	unTarCommand.Dir = unpackDir
	return RunExtractCommands(decompressCommand, unTarCommand)
}

func RunExtractCommands(decompressCommand *exec.Cmd, unTarCommand *exec.Cmd) error {
	decompressCommand.Stderr = os.Stderr
	decompressStdout, err := decompressCommand.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}

	unTarCommand.Stderr = os.Stderr
	unTarCommand.Stdin = decompressStdout

	return util.RunPipedCommands(decompressCommand, unTarCommand)
}

func GetCacheDirectory(appName string, envName string, isAvoidSystemOnWindows bool) (string, error) {
	env := os.Getenv(envName)
	if len(env) != 0 {
		return env, nil
	}

	currentOs := util.GetCurrentOs()
	if currentOs == util.MAC {
		userHomeDir, err := homedir.Dir()
		if err != nil {
			return "", errors.WithStack(err)
		}
		return filepath.Join(userHomeDir, "Library", "Caches", appName), nil
	}

	if currentOs == util.WINDOWS {
		localAppData := os.Getenv("LOCALAPPDATA")
		if len(localAppData) != 0 {
			// https://github.com/electron-userland/electron-builder/issues/1164
			if isAvoidSystemOnWindows && strings.Contains(strings.ToLower(localAppData), "\\windows\\system32\\") || strings.ToLower(os.Getenv("USERNAME")) == "system" {
				return filepath.Join(os.TempDir(), appName+"-cache"), nil
			}
			// https://github.com/sindresorhus/env-paths/blob/master/index.js
			return filepath.Join(localAppData, appName, "Cache"), nil
		}
	}

	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache != "" {
		return filepath.Join(xdgCache, appName), nil
	}

	userHomeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Join(userHomeDir, ".cache", appName), nil
}
