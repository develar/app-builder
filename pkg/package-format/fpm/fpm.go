package fpm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/pkg/errors"
)

type FpmConfiguration struct {
	Target string   `json:"target"`
	Args   []string `json:"args"`

	Compression string `json:"compression"`

	CustomDepends []string `json:"customDepends"`
	CustomRecommends []string `json:"customRecommends"`
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("fpm", "Build FPM targets.")

	configurationJson := command.Flag("configuration", "").Required().String()
	command.Action(func(context *kingpin.ParseContext) error {
		var configuration FpmConfiguration
		err := util.DecodeBase64IfNeeded(*configurationJson, &configuration)
		if err != nil {
			return err
		}

		var fpmPath string
		if util.GetCurrentOs() == util.WINDOWS || util.IsEnvTrue("USE_SYSTEM_FPM") {
			fpmPath = "fpm"
		} else {
			fpmDir, err := download.DownloadFpm()
			if err != nil {
				return err
			}
			fpmPath = filepath.Join(fpmDir, "fpm")
		}

		target := configuration.Target

		// must be first
		args := []string{"-s", "dir", "--force", "-t", target}
		if util.IsEnvTrue("FPM_DEBUG") {
			args = append(args, "--debug")
		}
		if log.IsDebugEnabled() {
			args = append(args, "--log", "debug")
		}
		args = configureDependencies(&configuration, target, args)
		args = configureRecommendations(&configuration, target, args)

		compression := "xz"
		if len(configuration.Compression) != 0 {
			compression = configuration.Compression
		}

		args = configureTargetSpecific(target, args, compression)

		args = append(args, configuration.Args...)

		command := exec.Command(fpmPath, args...)

		executablePath, err := os.Executable()
		if err != nil {
			return errors.WithStack(err)
		}

		env := os.Environ()
		env = append(env,
			"SZA_ARCHIVE_TYPE=xz",
			"FPM_COMPRESS_PROGRAM="+executablePath,
		)
		command.Env = env

		_, err = util.Execute(command)
		if err != nil {
			if execError, ok := err.(*util.ExecError); ok && strings.Contains(string(execError.Output), `"Need executable 'rpmbuild' to convert dir to rpm"`) {
				var installHint string
				if util.GetCurrentOs() == util.MAC {
					installHint = "brew install rpm"
				} else {
					installHint = "sudo apt-get install rpm"
				}
				log.LOG.Fatal("to build rpm, executable rpmbuild is required, please install: " + installHint)
			}
			return err
		}

		return nil
	})
}

func configureTargetSpecific(target string, args []string, compression string) []string {
	switch target {
	case "rpm":
		args = append(args, "--rpm-os", "linux")
		if compression == "xz" {
			args = append(args, "--rpm-compression", "xzmt")
		} else {
			args = append(args, "--rpm-compression", compression)
		}
	case "deb":
		args = append(args, "--deb-compression", compression)
	case "pacman":
		args = append(args, "--pacman-compression", compression)
	}
	return args
}

func configureDependencies(configuration *FpmConfiguration, target string, args []string) []string {
	depends := configuration.CustomDepends
	if len(depends) == 0 {
		depends = getDefaultDepends(target)
	}
	for _, value := range depends {
		args = append(args, "-d", value)
	}
	return args
}

func configureRecommendations(configuration *FpmConfiguration, target string, args []string) []string {
	if target == "deb" {
		recommends := configuration.CustomRecommends
		if len(recommends) == 0 {
			recommends = getDefaultRecommends(target)
		}
		for _, value := range recommends {
			args = append(args, "--deb-recommends", value)
		}
	}
	return args
}

//noinspection SpellCheckingInspection
func getDefaultDepends(target string) []string {
	switch target {
	case "deb":
		return []string{
			"libgtk-3-0", "libnotify4", "libnss3", "libxss1", "libxtst6", "xdg-utils", "libatspi2.0-0", "libuuid1", "libsecret-1-0",
		}

	case "rpm":
		return []string{
			"gtk3", /* for electron 2+ (electron 1 uses gtk2, but this old version is not supported anymore) */
			"libnotify", "nss", "libXScrnSaver", "libXtst", "xdg-utils",
			"at-spi2-core", /* since 5.0.0 */
			"libuuid",      /* since 4.0.0 */
		}

	case "pacman":
		return []string{"c-ares", "ffmpeg", "gtk3", "http-parser", "libevent", "libvpx", "libxslt", "libxss", "minizip", "nss", "re2", "snappy", "libnotify", "libappindicator-gtk3"}

	default:
		return nil
	}
}

func getDefaultRecommends(target string) []string {
	switch target {
	case "deb":
		return []string{
			"libappindicator3-1",
		}

	default:
		return nil
	}
}
