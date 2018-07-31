package electron

import (
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/go-fs-util"
)

func ConfigureUnpackCommand(app *kingpin.Application) {
	command := app.Command("unpack-electron", "")
	jsonConfig := command.Flag("configuration", "").Short('c').Required().String()
	outputDir := command.Flag("output", "").Required().String()
	distMacOsAppName := command.Flag("distMacOsAppName", "").Default("Electron.app").String()

	var cachedElectronZip string
	command.Action(func(context *kingpin.ParseContext) error {
		err := util.MapAsync(2, func(taskIndex int) (func() error, error) {
			if taskIndex == 0 {
				return func() error {
					return fsutil.EnsureEmptyDir(*outputDir)
				}, nil
			} else {
				return func() error {
					result, err := parseConfigAndDownload(jsonConfig)
					cachedElectronZip = result[0]
					return err
				}, nil
			}
		})
		if err != nil {
			return err
		}

		excludedFiles := make(map[string]bool)
		if download.GetCurrentOs() == download.MAC {
			excludedFiles[filepath.Join(*outputDir, *distMacOsAppName, "Contents", "Resources", "default_app.asar")] = true
		} else {
			excludedFiles[filepath.Join(*outputDir, "resources", "default_app.asar")] = true
		}
		excludedFiles[filepath.Join(*outputDir, "version")] = true

		err = Unzip(cachedElectronZip, *outputDir, excludedFiles)
		if err != nil {
			return err
		}

		return nil
	})
}
