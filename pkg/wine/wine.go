package wine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/util"
	"github.com/json-iterator/go"
	"github.com/mcuadros/go-version"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("wine", "")

	ia32Name := command.Flag("ia32", "The ia32 executable name").String()
	// x64Name not used for now
	_ = command.Flag("x64", "The x64 executable name").String()
	jsonEncodedArgs := command.Flag("args", "The json-encoded array of executable args").String()

	command.Validate(func(clause *kingpin.CmdClause) error {
		return nil
	})

	command.Action(func(context *kingpin.ParseContext) error {
		var parsedArgs []string
		if len(*jsonEncodedArgs) == 0 {
			parsedArgs = make([]string, 0)
		} else {
			err := jsoniter.UnmarshalFromString(*jsonEncodedArgs, &parsedArgs)
			if err != nil {
				return err
			}
		}

		return execWine(*ia32Name, parsedArgs)
	})
}

//noinspection GoUnusedParameter
func execWine(ia32Name string, args []string) error {
	args = append([]string{ia32Name}, args...)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if util.IsEnvTrue("USE_SYSTEM_WINE") {
		log.Debug("using system wine is forced")
	}

	if util.GetCurrentOs() == util.MAC {
		dirName := "wine-2.0.3-mac-10.13"
		checksum := "dlEVCf0YKP5IEiOKPNE48Q8NKXbXVdhuaI9hG2oyDEay2c+93PE5qls7XUbIYq4Xi1gRK8fkWeCtzN2oLpVQtg=="
		wineDir, err := download.DownloadArtifact(dirName, "https://github.com/electron-userland/electron-builder-binaries/releases/download/"+dirName+"/"+dirName+".7z", checksum)
		if err != nil {
			return err
		}

		command := exec.CommandContext(ctx, filepath.Join(wineDir, "bin/wine"), args...)
		env := os.Environ()
		env = append(env,
			fmt.Sprintf("WINEDEBUG=%s", "-all,err+all"),
			fmt.Sprintf("WINEDLLOVERRIDES=%s", "winemenubuilder.exe=d"),
			fmt.Sprintf("WINEPREFIX=%s", filepath.Join(wineDir, "wine-home")),
			fmt.Sprintf("DYLD_FALLBACK_LIBRARY_PATH=%s", filepath.Join(wineDir, "lib")+":"+os.Getenv("DYLD_FALLBACK_LIBRARY_PATH")),
		)
		command.Env = env
		_, err = util.Execute(command, "")
		if err != nil {
			return err
		}

		return nil
	}

	err := checkWineVersion()
	if err != nil {
		return err
	}

	_, err = util.Execute(exec.CommandContext(ctx, "wine", args...), "")
	if err != nil {
		return err
	}

	return nil
}

func checkWineVersion() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	wineVersionResult, err := exec.CommandContext(ctx, "wine", "--version").Output()
	if err != nil {
		log.WithError(err).Debug("wine version check result")
		return util.NewMessageError("wine is required, please see https://electron.build/multi-platform-build#linux", "ERR_WINE_NOT_INSTALLED")
	}
	return doCheckWineVersion(strings.TrimPrefix(strings.TrimSpace(string(wineVersionResult)), "wine-"))
}

func doCheckWineVersion(wineVersion string) error {
	spaceIndex := strings.IndexRune(wineVersion, ' ')
	if spaceIndex > 0 {
		wineVersion = wineVersion[0:spaceIndex]
	}

	suffixIndex := strings.IndexRune(wineVersion, '-')
	if suffixIndex > 0 {
		wineVersion = wineVersion[0:suffixIndex]
	}

	if version.Compare(wineVersion, "1.8.0", "<") {
		return util.NewMessageError(`wine 1.8+ is required, but your version is `+wineVersion, "ERR_WINE_VERSION_INCOMPATIBLE")
	}
	return nil
}
