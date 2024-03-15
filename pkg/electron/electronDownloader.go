package electron

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

type ElectronDownloadOptions struct {
	Version  string `json:"version"`
	CacheDir string `json:"cache"`
	Mirror   string `json:"mirror"`

	Platform string `json:"platform"`
	Arch     string `json:"arch"`

	CustomDir      string `json:"customDir"`
	CustomFilename string `json:"customFilename"`
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("download-electron", "")
	jsonConfig := command.Flag("configuration", "").Short('c').Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		configs, err := parseConfig(jsonConfig)
		if err != nil {
			return err
		}

		_, err = downloadElectron(configs)
		return err
	})
}

func parseConfig(jsonConfig *string) ([]ElectronDownloadOptions, error) {
	var configs []ElectronDownloadOptions
	err := jsoniter.UnmarshalFromString(*jsonConfig, &configs)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func downloadElectron(configs []ElectronDownloadOptions) ([]string, error) {
	result := make([]string, len(configs))
	return result, util.MapAsync(len(configs), func(taskIndex int) (func() error, error) {
		config := configs[taskIndex]
		return func() error {
			cacheDir := config.CacheDir
			if cacheDir == "" {
				var err error
				cacheDir, err = download.GetCacheDirectory("electron", "ELECTRON_CACHE", false)
				if err != nil {
					return err
				}
			}

			electronDownloader := &ElectronDownloader{
				config:   &config,
				cacheDir: cacheDir,
			}

			cachedFile, err := electronDownloader.Download()
			if err != nil {
				return err
			}

			result[taskIndex] = cachedFile

			return nil
		}, nil
	})
}

func getBaseUrl(config *ElectronDownloadOptions) string {
	v := config.Mirror
	if len(v) == 0 {
		v = os.Getenv("NPM_CONFIG_ELECTRON_MIRROR")
	}
	if len(v) == 0 {
		v = os.Getenv("npm_config_electron_mirror")
	}
	if len(v) == 0 {
		v = os.Getenv("ELECTRON_MIRROR")
	}
	if len(v) == 0 {
		if strings.Contains(config.Version, "-nightly.") {
			v = "https://github.com/electron/nightlies/releases/download/"
		} else {
			v = "https://github.com/electron/electron/releases/download/"
		}
	}
	// Compatibility with previous code caused user who need to set mirror with a suffix `/v`
	if strings.HasSuffix(v, "/v") {
		v = v[:len(v)-1]
	}
	return v
}

func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func getMiddleUrl(config *ElectronDownloadOptions) string {
	v := os.Getenv("ELECTRON_CUSTOM_DIR")
	if len(v) == 0 {
		v = config.CustomDir
	}
	if len(v) == 0 {
		v = normalizeVersion(config.Version)
	}
	return v
}

func getUrlSuffix(config *ElectronDownloadOptions) string {
	v := os.Getenv("ELECTRON_CUSTOM_FILENAME")
	if len(v) == 0 {
		v = config.CustomFilename
	}
	if len(v) == 0 {
		v = getFilename(config)
	}
	return v
}

func getFilename(config *ElectronDownloadOptions) string {
	return "electron-" + normalizeVersion(config.Version) + "-" + config.Platform + "-" + config.Arch + ".zip"
}

type ElectronDownloader struct {
	config *ElectronDownloadOptions

	cacheDir string
}

func (t *ElectronDownloader) getCachedFile() string {
	fileName := t.config.CustomFilename
	if len(fileName) == 0 {
		fileName = getFilename(t.config)
	}
	return filepath.Join(t.cacheDir, fileName)
}

func (t *ElectronDownloader) Download() (string, error) {
	if t.config.Version == "" {
		return "", errors.New("version not specified")
	}
	if t.config.Platform == "" {
		return "", errors.New("platform not specified")
	}
	if t.config.Arch == "" {
		return "", errors.New("arch not specified")
	}

	cachedFile := t.getCachedFile()

	fileInfo, err := os.Stat(cachedFile)
	if err != nil && !os.IsNotExist(err) {
		return "", errors.WithStack(err)
	}

	if fileInfo != nil {
		if fileInfo.IsDir() {
			return "", errors.New("File expected, but got dir")
		}
		return cachedFile, nil
	}

	err = fsutil.EnsureDir(t.cacheDir)
	if err != nil {
		return "", errors.WithStack(err)
	}

	url := getBaseUrl(t.config) + getMiddleUrl(t.config) + "/" + getUrlSuffix(t.config)
	err = t.doDownload(url, cachedFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return cachedFile, nil
}

func (t *ElectronDownloader) doDownload(url string, cachedFile string) error {
	tempFile, err := util.TempFile(t.cacheDir, ".zip")
	if err != nil {
		return errors.WithStack(err)
	}

	downloader := download.NewDownloader()
	err = downloader.Download(url, tempFile, "")
	if err != nil {
		return errors.WithStack(err)
	}

	download.RenameToFinalFile(tempFile, cachedFile, log.LOG.With(zap.String("url", url), zap.String("path", cachedFile)))
	return nil
}
