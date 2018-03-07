package appimage

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/blockmap"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/linuxTools"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
)

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

	isRemoveStage := util.ConfigureIsRemoveStageParam(command)

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

	appImageToolDir, err := linuxTools.GetAppImageToolDir()
	if err != nil {
		return errors.WithStack(err)
	}

	arch := *options.arch
	if arch == "x64" || arch == "ia32" {
		err = fs.CopyUsingHardlink(filepath.Join(appImageToolDir, "lib", arch), filepath.Join(stageDir, "usr", "lib"))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	runtimeData, err := ioutil.ReadFile(filepath.Join(appImageToolDir, "runtime-"+arch))
	if err != nil {
		return errors.WithStack(err)
	}

	err = createSquashFs(options, len(runtimeData))
	if err != nil {
		return errors.WithStack(err)
	}

	outputFile := *options.output
	err = writeRuntimeData(outputFile, runtimeData)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Chmod(outputFile, 0755)
	if err != nil {
		return errors.WithStack(err)
	}

	updateInfo, err := blockmap.BuildBlockMap(outputFile, blockmap.DefaultChunkerConfiguration, blockmap.DEFLATE, "")
	if err != nil {
		return errors.WithStack(err)
	}

	err = util.WriteJsonToStdOut(updateInfo)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func writeRuntimeData(filePath string, runtimeData []byte) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = file.WriteAt(runtimeData, 0)
	return util.CloseAndCheckError(err, file)
}

func createSquashFs(options AppImageOptions, offset int) error {
	mksquashfsPath, err := linuxTools.GetMksquashfs()
	if err != nil {
		return errors.WithStack(err)
	}

	var args []string
	args = append(args, *options.stageDir, *options.output, "-offset", strconv.Itoa(offset), "-all-root", "-noappend", "-no-progress", "-quiet", "-no-xattrs", "-no-fragments")
	// "-mkfs-fixed-time", "0" not available for mac yet (since AppImage developers don't provide actual version of mksquashfs for macOS and no official mksquashfs build for macOS)
	if *options.compression != "" {
		// default gzip compression - 51.9, xz - 50.4 difference is negligible, start time - well, it seems, a little bit longer (but on Parallels VM on external SSD disk)
		// so, to be decided later, is it worth to use xz by default
		args = append(args, "-comp", *options.compression)
		if *options.compression == "xz" {
			//noinspection SpellCheckingInspection
			args = append(args, "-Xdict-size", "100%", "-b", "1048576")
		}
	}

	err = util.Execute(exec.Command(mksquashfsPath, args...), *options.stageDir)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}