package download

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/develar/app-builder/util"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// we cache in the global location - in the home dir, not in the node_modules/.cache (https://www.npmjs.com/package/find-cache-dir) because
// * don't need to find node_modules
// * don't pollute user project dir (important in case of 1-package.json project structure)
// * simplify/speed-up tests (don't download fpm for each test project)
func DownloadArtifact(dirName string, url string, checksum string) (string, error) {
	cacheDir, err := getCacheDirectory("electron-builder")
	if err != nil {
		return "", errors.WithStack(err)
	}

	cachePath := filepath.Join(cacheDir, dirName[0:strings.Index(dirName, "-")])
	dirPath := filepath.Join(cachePath, dirName)

	logFields := log.Fields{
		"path": dirPath,
	}

	dirStat, err := os.Stat(dirPath)
	if err == nil && dirStat.IsDir() {
		log.WithFields(logFields).Debug("found existing")
		return dirPath, nil
	}

	if err != nil && !os.IsNotExist(err) {
		return "", errors.WithMessage(err, "error during cache check for path "+dirPath)
	}

	err = os.MkdirAll(cachePath, 0700)
	if err != nil {
		return "", errors.WithStack(err)
	}

	log.WithFields(logFields).WithField("url", url).Info("downloading")

	// 7z cannot be extracted from the input stream, temp file is required
	tempUnpackDir, err := util.TempDir(cachePath, "")
	if err != nil {
		return "", errors.WithStack(err)
	}

	archiveName := tempUnpackDir + ".7z"
	err = Download(url, archiveName, checksum)
	if err != nil {
		return "", errors.WithStack(err)
	}

	command := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), "x", "-bd", archiveName, "-o"+tempUnpackDir)
	command.Dir = cachePath
	output, err := command.CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Debug(string(output))
	err = os.Remove(archiveName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = os.Rename(tempUnpackDir, dirPath)
	if err != nil {
		log.WithFields(logFields).WithFields(log.Fields{
			"tempUnpackDir": tempUnpackDir,
			"error":         err,
		}).Warn("cannot move downloaded into final location (another process downloaded faster?)")
	}

	log.WithFields(logFields).Debug("downloaded")

	return dirPath, nil
}

func getCacheDirectory(dirName string) (string, error) {
	env := os.Getenv("ELECTRON_BUILDER_CACHE")
	if len(env) != 0 {
		return env, nil
	}

	if runtime.GOOS == "darwin" {
		userHomeDir, err := homedir.Dir()
		if err != nil {
			return "", errors.WithStack(err)
		}
		return filepath.Join(userHomeDir, "Library", "Caches", dirName), nil
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if runtime.GOOS == "windows" && len(localAppData) != 0 {
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
