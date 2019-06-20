package main

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/archive/zipx"
	"github.com/develar/app-builder/pkg/blockmap"
	"github.com/develar/app-builder/pkg/codesign"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/electron"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/icons"
	"github.com/develar/app-builder/pkg/linuxTools"
	"github.com/develar/app-builder/pkg/log-cli"
	"github.com/develar/app-builder/pkg/node-modules"
	"github.com/develar/app-builder/pkg/package-format/appimage"
	"github.com/develar/app-builder/pkg/package-format/dmg"
	"github.com/develar/app-builder/pkg/package-format/proton-native"
	"github.com/develar/app-builder/pkg/package-format/snap"
	"github.com/develar/app-builder/pkg/publisher"
	"github.com/develar/app-builder/pkg/remoteBuild"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/app-builder/pkg/wine"
	"github.com/develar/errors"
	"github.com/segmentio/ksuid"
)

func main() {
	log_cli.InitLogger()

	if os.Getenv("SZA_ARCHIVE_TYPE") != "" {
		err := compress()
		if err != nil {
			util.LogErrorAndExit(err)
		}
		return
	}

	var app = kingpin.New("app-builder", "app-builder").Version("2.6.18")

	node_modules.ConfigureCommand(app)
	//codesign.ConfigureCommand(app)
	publisher.ConfigurePublishToS3Command(app)
	remoteBuild.ConfigureBuildCommand(app)

	download.ConfigureCommand(app)
	download.ConfigureArtifactCommand(app)

	electron.ConfigureCommand(app)
	electron.ConfigureUnpackCommand(app)

	zipx.ConfigureUnzipCommand(app)
	proton_native.ConfigureCommand(app)

	configurePrefetchToolsCommand(app)

	ConfigureCopyCommand(app)
	appimage.ConfigureCommand(app)
	snap.ConfigureCommand(app)

	err := icons.ConfigureCommand(app)
	if err != nil {
		util.LogErrorAndExit(err)
	}

	dmg.ConfigureCommand(app)
	blockmap.ConfigureCommand(app)
	codesign.ConfigureCertificateInfoCommand(app)

	wine.ConfigureCommand(app)
	configureKsUidCommand(app)

	_, err = app.Parse(os.Args[1:])
	if err != nil {
		util.LogErrorAndExit(err)
	}
}

func ConfigureCopyCommand(app *kingpin.Application) {
	command := app.Command("copy", "Copy file or dir.")
	from := command.Flag("from", "").Required().Short('f').String()
	to := command.Flag("to", "").Required().Short('t').String()
	isUseHardLinks := command.Flag("hard-link", "Whether to use hard-links if possible").Bool()

	command.Action(func(context *kingpin.ParseContext) error {
		var fileCopier fs.FileCopier
		fileCopier.IsUseHardLinks = *isUseHardLinks
		return errors.WithStack(fileCopier.CopyDirOrFile(*from, *to))
	})
}

func configureKsUidCommand(app *kingpin.Application) {
	command := app.Command("ksuid", "Generate KSUID")
	command.Action(func(context *kingpin.ParseContext) error {
		_, err := os.Stdout.Write([]byte(ksuid.New().String()))
		return errors.WithStack(err)
	})
}

func compress() error {
	args := []string{"a", "-si", "-so", "-t" + util.GetEnvOrDefault("SZA_ARCHIVE_TYPE", "xz"), "-mx" + util.GetEnvOrDefault("SZA_COMPRESSION_LEVEL", "9"), "dummy"}
	args = append(args, os.Args[1:]...)

	command := exec.Command(util.Get7zPath(), args...)
	command.Stderr = os.Stderr

	stdin, err := command.StdinPipe()
	if nil != err {
		return errors.WithStack(err)
	}

	stdout, err := command.StdoutPipe()
	if nil != err {
		return errors.WithStack(err)
	}

	err = command.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		defer util.Close(stdin)
		_, _ = io.Copy(stdin, os.Stdin)
	}()

	go func() {
		defer waitGroup.Done()
		_, _ = io.Copy(os.Stdout, stdout)
	}()

	waitGroup.Wait()
	err = command.Wait()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func configurePrefetchToolsCommand(app *kingpin.Application) {
	command := app.Command("prefetch-tools", "Prefetch all required tools")
	osName := command.Flag("osName", "").Default(runtime.GOOS).Enum("darwin", "linux", "win32")
	command.Action(func(context *kingpin.ParseContext) error {
		_, err := linuxTools.GetAppImageToolDir()
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = snap.ResolveTemplateFile("", "electron4", "")
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = download.DownloadFpm()
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = download.DownloadZstd(util.ToOsName(*osName))
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
}
