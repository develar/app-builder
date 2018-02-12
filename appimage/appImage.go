package appimage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/download"
	"github.com/develar/app-builder/fs"
	"github.com/develar/app-builder/util"
	"github.com/develar/errors"
)

//noinspection GoSnakeCaseUsage,SpellCheckingInspection
const APPIMAGE_TOOL_SHA512 = "XXktBh+pgIUvpaIPjLmt03kmqnMWopkCSVrGm9zou1Jg+1HjXMhvqnUSwHxi4QUpPnblp05fJXSsxgJXNOjKog=="

type AppImageOptions struct {
	appDir   *string
	stageDir *string
	arch     *string
	output   *string

	compression *string
}

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("appimage", "Build AppImage.")

	options := AppImageOptions{
		appDir:   command.Flag("app", "The app dir.").Short('a').Required().String(),
		stageDir: command.Flag("stage", "The stage dir.").Short('s').Required().String(),
		output:   command.Flag("output", "The output file.").Short('o').Required().String(),
		arch:     command.Flag("arch", "The arch.").Default("x64").Enum("x64", "ia32", "armv7l", "arm64"),

		compression: command.Flag("compression", "The compression.").Enum("xz", "gzip"),
	}

	isRemoveStage := command.Flag("remove-stage", "Whether to remove stage after build.").Default("true").Bool()

	command.Action(func(context *kingpin.ParseContext) error {
		err := AppImage(options)
		if err != nil {
			return errors.WithStack(err)
		}

		if *isRemoveStage {
			err = os.RemoveAll(*options.stageDir)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	})
}

func AppImage(options AppImageOptions) error {
	stageDir := *options.stageDir

	err := fs.CopyUsingHardlink(*options.appDir, filepath.Join(stageDir, "app"))
	if err != nil {
		return errors.WithStack(err)
	}

	dirName := "appimage-9.0.6"
	appImageToolDir, err := download.DownloadArtifact("", "https://github.com/electron-userland/electron-builder-binaries/releases/download/"+dirName+"/"+dirName+".7z", APPIMAGE_TOOL_SHA512)
	if err != nil {
		return errors.WithStack(err)
	}

	arch := *options.arch
	if arch == "x64" || arch == "ia32" {
		err = fs.CopyUsingHardlink(filepath.Join(appImageToolDir, "lib", arch), filepath.Join(stageDir, "usr", "lib"))
		if err != nil {
			return err
		}
	}

	var args []string
	args = append(args, "--runtime-file", filepath.Join(appImageToolDir, "runtime-"+arch), "--no-appstream")
	if *options.compression != "" {
		// default gzip compression - 51.9, xz - 50.4 difference is negligible, start time - well, it seems, a little bit longer (but on Parallels VM on external SSD disk)
		// so, to be decided later, is it worth to use xz by default
		args = append(args, "--comp", *options.compression)
	}
	args = append(args, stageDir, *options.output)

	var vendorToolDir string
	if runtime.GOOS == "darwin" {
		vendorToolDir = filepath.Join(appImageToolDir, "darwin")

	} else {
		vendorToolDir = filepath.Join(appImageToolDir, "linux-"+goArchToNodeArch(runtime.GOARCH))
	}

	command := exec.Command(filepath.Join(vendorToolDir, "appimagetool"), args...)

	appImageArch, err := toAppImageArch(arch)
	if err != nil {
		return err
	}

	env := os.Environ()
	env = append(env,
		fmt.Sprintf("PATH=%s", vendorToolDir+":"+os.Getenv("PATH")),
		// to avoid detection by appimagetool (see extract_arch_from_text about expected arch names)
		fmt.Sprintf("ARCH=%s", appImageArch),
	)
	command.Env = env

	err = util.Execute(command, stageDir)
	if err != nil {
		return err
	}
	return nil
}

func toAppImageArch(arch string) (string, error) {
	switch arch {
	case "x64":
		return "x86_64", nil
	case "ia32":
		return "i386", nil
	case "armv7l":
		return "arm", nil
	case "arm64":
		return "arm_aarch64", nil

	default:
		return "", errors.Errorf("unsupported arch %s", arch)
	}
}

func goArchToNodeArch(arch string) (string) {
	switch arch {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	default:
		return arch
	}
}
