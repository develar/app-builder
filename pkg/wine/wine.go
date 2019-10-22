package wine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/json-iterator/go"
	"github.com/mcuadros/go-version"
	"go.uber.org/zap"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("wine", "")

	ia32Name := command.Flag("ia32", "The ia32 executable name").String()
	// x64Name not used for now
	x64Name := command.Flag("x64", "The x64 executable name").String()
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

		return ExecWine(*ia32Name, *x64Name, parsedArgs)
	})
}

func isMacOsCatalina() (bool, error) {
	osRelease, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false, err
	}

	return version.Compare(strings.TrimSpace(string(osRelease)), "19.0.0", ">="), nil
}

//noinspection GoUnusedParameter
func ExecWine(ia32Name string, ia64Name string, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	useSystemWine := util.IsEnvTrue("USE_SYSTEM_WINE")
	if useSystemWine {
		log.Debug("using system wine is forced")
	}

	if util.GetCurrentOs() == util.MAC {
		return executeMacOsWine(useSystemWine, ctx, args, ia32Name, ia64Name)
	}

	err := checkWineVersion()
	if err != nil {
		return err
	}

	args = append([]string{ia32Name}, args...)
	_, err = util.Execute(exec.CommandContext(ctx, "wine", args...))
	if err != nil {
		return err
	}

	return nil
}

func executeMacOsWine(useSystemWine bool, ctx context.Context, args []string, ia32Name string, ia64Name string) error {
	catalina, err := isMacOsCatalina()
	if err != nil {
		log.Warn("cannot detect macOS version", zap.Error(err))
	}

	if catalina {
		if len(ia64Name) == 0 {
			return errors.New("macOS Catalina doesn't support 32-bit executables and as result Wine cannot run Windows 32-bit applications too")
		}

		args = append([]string{ia64Name}, args...)
	} else {
		args = append([]string{ia32Name}, args...)
	}

	if useSystemWine {
		command := exec.CommandContext(ctx, "wine", args...)
		env := os.Environ()
		env = append(env,
			fmt.Sprintf("WINEDEBUG=%s", "-all,err+all"),
			fmt.Sprintf("WINEDLLOVERRIDES=%s", "winemenubuilder.exe=d"),
		)
		command.Env = env

		if _, err := util.Execute(command); err != nil {
			return err
		}

		return nil
	}

	var wineDir string
	var wineExecutable string
	if catalina {
		dirName := "wine-4.0.1-mac"
		//noinspection SpellCheckingInspection
		checksum := "aCUQOyuPGlEvLMp0lPzb54D96+8IcLwmKTMElrZZqVWtEL1LQC7L9XpPv4RqaLX3BOeSifneEi4j9DpYdC1DCA=="
		wineDir, err = download.DownloadArtifact(dirName, download.GetGithubBaseUrl()+dirName+"/"+dirName+".7z", checksum)
		if err != nil {
			return err
		}

		wineExecutable = "wine64"
	} else {
		dirName := "wine-2.0.3-mac-10.13"
		//noinspection SpellCheckingInspection
		checksum := "dlEVCf0YKP5IEiOKPNE48Q8NKXbXVdhuaI9hG2oyDEay2c+93PE5qls7XUbIYq4Xi1gRK8fkWeCtzN2oLpVQtg=="
		wineDir, err = download.DownloadArtifact(dirName, download.GetGithubBaseUrl()+dirName+"/"+dirName+".7z", checksum)
		if err != nil {
			return err
		}

		wineExecutable = "wine"
	}
	command := exec.CommandContext(ctx, filepath.Join(wineDir, "bin", wineExecutable), args...)
	env := os.Environ()
	//noinspection SpellCheckingInspection
	env = append(env,
		fmt.Sprintf("WINEDEBUG=%s", "-all,err+all"),
		fmt.Sprintf("WINEDLLOVERRIDES=%s", "winemenubuilder.exe=d"),
		"WINEPREFIX=" + filepath.Join(wineDir, "wine-home"),
		fmt.Sprintf("DYLD_FALLBACK_LIBRARY_PATH=%s", filepath.Join(wineDir, "lib")+":"+os.Getenv("DYLD_FALLBACK_LIBRARY_PATH")),
	)

	//if catalina && len(ia64Name) == 0 {
	//	//noinspection SpellCheckingInspection
	//	env = append(env,
	//		"WINEARCH=win32",
	//		"WINEPREFIX="+filepath.Join(wineDir, "wine-home-ia32"),
	//	)
	//} else {
	env = append(env, "WINEPREFIX="+filepath.Join(wineDir, "wine-home"))
	//}
	command.Env = env
	_, err = util.Execute(command)
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
		log.Debug("wine version check result", zap.Error(err))
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
