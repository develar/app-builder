package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/asar"
	"github.com/develar/app-builder/blockmap"
	"github.com/develar/app-builder/download"
	"github.com/develar/app-builder/errors"
	"github.com/develar/app-builder/fs"
	"github.com/develar/app-builder/icons"
	"github.com/develar/app-builder/log-cli"
	"github.com/develar/app-builder/snap"
	"github.com/develar/app-builder/util"
)

var (
	appVersion = "1.2.1"
	app        = kingpin.New("app-builder", "app-builder").Version(appVersion)

	convertIcon          = app.Command("icon", "create ICNS or ICO or icon set from PNG files")
	convertIconSources   = convertIcon.Flag("input", "input directory or file").Short('i').Required().Strings()
	convertIconOutFormat = convertIcon.Flag("format", "output format").Short('f').Required().Enum("icns", "ico", "set")
	convertIconRoots     = convertIcon.Flag("root", "base directory to resolve relative path").Short('r').Strings()

	buildBlockMap            = app.Command("blockmap", "Generates file block map for differential update using content defined chunking (that is robust to insertions, deletions, and changes to input file)")
	buildBlockMapInFile      = buildBlockMap.Flag("input", "input file").Short('i').Required().String()
	buildBlockMapOutFile     = buildBlockMap.Flag("output", "output file").Short('o').String()
	buildBlockMapCompression = buildBlockMap.Flag("compression", "compression, one of: gzip, deflate").Short('c').Default("gzip").Enum("gzip", "deflate")

	buildAsar        = app.Command("asar", "")
	buildAsarOutFile = buildAsar.Flag("output", "").Required().String()
)

func main() {
	download.ConfigureCommand(app)
	download.ConfigureArtifactCommand(app)
	ConfigureCopyCommand(app)
	snap.ConfigureCommand(app)

	log.SetHandler(log_cli.Default)

	debugEnv, isDebugDefined := os.LookupEnv("DEBUG")
	if isDebugDefined && debugEnv != "false" {
		log.SetLevel(log.DebugLevel)
	}

	if os.Getenv("SZA_ARCHIVE_TYPE") != "" {
		err := compress()
		if err != nil {
			errors.LogErrorAndExit(err)
		}
		return
	}

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		errors.LogErrorAndExit(err)
	}

	switch command {
	case convertIcon.FullCommand():
		doConvertIcon()

	case buildAsar.FullCommand():
		err := asar.BuildAsar(*buildAsarOutFile)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case buildBlockMap.FullCommand():
		err := doBuildBlockMap()
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case buildBlockMap.FullCommand():
		err := doBuildBlockMap()
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
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
		return errors.WithStack(fileCopier.CopyDirOrFile(*to, *from))
	})
}

func compress() error {
	args := []string{"a", "-si", "-so", "-t" + util.GetEnvOrDefault("SZA_ARCHIVE_TYPE", "xz"), "-mx" + util.GetEnvOrDefault("SZA_COMPRESSION_LEVEL", "9"), "dummy"}
	args = append(args, os.Args[1:]...)

	command := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), args...)
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
		defer stdin.Close()
		io.Copy(stdin, os.Stdin)
	}()

	go func() {
		defer waitGroup.Done()
		io.Copy(os.Stdout, stdout)
	}()

	waitGroup.Wait()
	err = command.Wait()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func doConvertIcon() {
	resultFile, err := icons.ConvertIcon(*convertIconSources, *convertIconRoots, *convertIconOutFormat)
	if err != nil {
		log.Debugf("%+v\n", err)

		switch t := errors.Cause(err).(type) {
		default:
			log.Fatalf("%+v\n", err)
			return

		case *icons.ImageSizeError:
			printAppError(t)
			return

		case *icons.ImageFormatError:
			printAppError(t)
			return
		}
	}

	err = writeJsonToStdOut(icons.IconConvertResult{Icons: resultFile})
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func printAppError(error icons.ImageError) {
	err := writeJsonToStdOut(icons.MisConfigurationError{Message: error.Error(), Code: error.ErrorCode()})
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func doBuildBlockMap() error {
	var compressionFormat blockmap.CompressionFormat
	switch *buildBlockMapCompression {
	case "gzip":
		compressionFormat = blockmap.GZIP
	case "deflate":
		compressionFormat = blockmap.DEFLATE
	default:
		return fmt.Errorf("unknown compression format %s", *buildBlockMapCompression)
	}

	inputInfo, err := blockmap.BuildBlockMap(*buildBlockMapInFile, blockmap.DefaultChunkerConfiguration, compressionFormat, *buildBlockMapOutFile)
	if err != nil {
		return err
	}

	return writeJsonToStdOut(inputInfo)
}

func writeJsonToStdOut(v interface{}) error {
	serializedInputInfo, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(serializedInputInfo)
	if err != nil {
		return err
	}

	return nil
}
