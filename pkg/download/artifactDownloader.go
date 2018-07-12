package download

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/mitchellh/go-homedir"
)

type osName int

const (
	MAC osName = iota
	LINUX
	WINDOWS
)

//noinspection GoExportedFuncWithUnexportedType
func GetCurrentOs() osName {
	return ToOsName(runtime.GOOS)
}

//noinspection GoExportedFuncWithUnexportedType
func ToOsName(name string) osName {
	if name == "windows" || name == "win32" {
		return WINDOWS
	} else if name == "darwin" {
		return MAC
	} else {
		return LINUX
	}
}

func ConfigureArtifactCommand(app *kingpin.Application) {
	command := app.Command("download-artifact", "Download, unpack and cache artifact from GitHub.")
	name := command.Flag("name", "The artifact name.").Short('n').Required().String()
	url := command.Flag("url", "The artifact URL.").Short('u').Required().String()
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

func getCacheDirectoryForArtifact(dirName string) (string, error) {
	result, err := GetCacheDirectory("electron-builder")
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

// we cache in the global location - in the home dir, not in the node_modules/.cache (https://www.npmjs.com/package/find-cache-dir) because
// * don't need to find node_modules
// * don't pollute user project dir (important in case of 1-package.json project structure)
// * simplify/speed-up tests (don't download fpm for each test project)
func DownloadArtifact(dirName string, url string, checksum string) (string, error) {
	if dirName == "fpm" {
		return DownloadFpm()
	} else if dirName == "zstd" {
		return DownloadZstd(GetCurrentOs())
	}

	isNodeJsArtifact := dirName == "node"
	if isNodeJsArtifact {
		versionAndArch := url
		version := versionAndArch[0:strings.Index(versionAndArch, "-")]
		url = "https://nodejs.org/dist/v" + version + "/node-v" + versionAndArch + ".tar.xz"
		dirName = dirName + "-" + versionAndArch
	} else if len(dirName) == 0 {
		dirName = path.Base(url)
		// cannot simply find fist dot because file name can contains version like 9.1.0
		dirName = strings.TrimSuffix(dirName, ".7z")
		dirName = strings.TrimSuffix(dirName, ".tar")
	}

	cacheDir, err := getCacheDirectoryForArtifact(dirName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	filePath := filepath.Join(cacheDir, dirName)

	logFields := log.Fields{
		"path": filePath,
	}

	dirStat, err := os.Stat(filePath)
	if err == nil && dirStat.IsDir() {
		log.WithFields(logFields).Debug("found existing")
		return filePath, nil
	}

	if err != nil && !os.IsNotExist(err) {
		return "", errors.WithMessage(err, "error during cache check for path "+filePath)
	}

	err = os.MkdirAll(cacheDir, 0777)
	if err != nil {
		return "", errors.WithStack(err)
	}

	log.WithFields(logFields).WithField("url", url).Info("downloading")

	// 7z cannot be extracted from the input stream, temp file is required
	tempUnpackDir, err := util.TempDir(cacheDir, "")
	if err != nil {
		return "", errors.WithStack(err)
	}

	var archiveName string
	if isNodeJsArtifact {
		archiveName = tempUnpackDir + ".tar.xz"
	} else {
		archiveName = tempUnpackDir + ".7z"
	}

	err = NewDownloader().Download(url, archiveName, checksum)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if isNodeJsArtifact {
		err = unpackTarXzNodeJs(archiveName, tempUnpackDir)
		if err != nil {
			return "", errors.WithStack(err)
		}
	} else if strings.HasSuffix(url, ".tar.7z") {
		err = unpackTar7z(archiveName, tempUnpackDir)
		if err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		command := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), "x", "-bd", archiveName, "-o"+tempUnpackDir)
		command.Dir = cacheDir
		output, err := command.CombinedOutput()
		if err != nil {
			return "", errors.WithStack(err)
		}

		log.Debug(string(output))
	}

	err = os.Remove(archiveName)
	if err != nil {
		log.WithFields(logFields).WithFields(log.Fields{
			"tempUnpackDir": tempUnpackDir,
			"error":         err,
		}).Warn("cannot remove downloaded archive (another process downloaded faster?)")
	}

	err = os.Rename(tempUnpackDir, filePath)
	if err != nil {
		log.WithFields(logFields).WithFields(log.Fields{
			"tempUnpackDir": tempUnpackDir,
			"error":         err,
		}).Warn("cannot move downloaded into final location (another process downloaded faster?)")
	}

	log.WithFields(logFields).Debug("downloaded")

	return filePath, nil
}

func unpackTarXzNodeJs(archiveName string, unpackDir string) error {
	decompressCommand := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), "e", "-bd", "-txz", archiveName, "-so")

	//noinspection SpellCheckingInspection
	unTarCommand := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), "e", "-bd", "-ttar", "-o"+unpackDir, "*/bin/node", "-r", "-si")
	err := runExtractCommands(decompressCommand, unTarCommand)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Chmod(filepath.Join(unpackDir, "node"), 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func unpackTar7z(archiveName string, unpackDir string) error {
	decompressCommand := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), "e", "-bd", "-t7z", archiveName, "-so")

	args := []string{"-x"}
	if runtime.GOOS == "darwin" {
		// otherwise snap error review "unusual mode 'rwxr-xr-x' for symlink"
		args = append(args, "-p")
	}
	args = append(args, "-f", "-")

	//noinspection SpellCheckingInspection
	unTarCommand := exec.Command("tar", args...)
	unTarCommand.Dir = unpackDir
	return runExtractCommands(decompressCommand, unTarCommand)
}

func runExtractCommands(decompressCommand *exec.Cmd, unTarCommand *exec.Cmd) error {
	decompressCommand.Stderr = os.Stderr
	decompressStdout, err := decompressCommand.StdoutPipe()
	if err != nil {
		return errors.WithStack(err)
	}

	unTarCommand.Stderr = os.Stderr
	unTarCommand.Stdin = decompressStdout

	return util.RunPipedCommands(decompressCommand, unTarCommand)
}

func GetCacheDirectory(dirName string) (string, error) {
	env := os.Getenv("ELECTRON_BUILDER_CACHE")
	if len(env) != 0 {
		return env, nil
	}

	currentOs := GetCurrentOs()
	if currentOs == MAC {
		userHomeDir, err := homedir.Dir()
		if err != nil {
			return "", errors.WithStack(err)
		}
		return filepath.Join(userHomeDir, "Library", "Caches", dirName), nil
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if currentOs == WINDOWS && len(localAppData) != 0 {
		// https://github.com/electron-userland/electron-builder/issues/1164
		if strings.Contains(strings.ToLower(localAppData), "\\windows\\system32\\") || strings.ToLower(os.Getenv("USERNAME")) == "system" {
			return filepath.Join(os.TempDir(), dirName+"-cache"), nil
		}
		return filepath.Join(localAppData, dirName, "cache"), nil
	}

	userHomeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Join(userHomeDir, ".cache", "electron-builder"), nil
}