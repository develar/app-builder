package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/asar"
	"github.com/develar/app-builder/blockmap"
	"github.com/develar/app-builder/download"
	"github.com/develar/app-builder/icons"
	"github.com/develar/app-builder/log-cli"
	"github.com/develar/app-builder/util"
	"github.com/pkg/errors"
)

var (
	appVersion = "1.1.0"
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

	copyDirCommand     = app.Command("copy", "")
	copyDirSource      = copyDirCommand.Flag("from", "").Required().Short('f').String()
	copyDirDestination = copyDirCommand.Flag("to", "").Required().Short('t').String()

	//cleanupSnapCommand = app.Command("clean-snap", "")
	//cleanupSnapCommandDir = cleanupSnapCommand.Flag("dir", "").Required().String()

	downloadCommand         = app.Command("download", "")
	downloadCommandUrl      = downloadCommand.Flag("url", "The URL").Short('u').Required().String()
	downloadCommandOutput   = downloadCommand.Flag("output", "The output file").Short('o').Required().String()
	downloadCommandChecksum = downloadCommand.Flag("sha512", "The expected sha512 of file").String()

	downloadArtifactCommand         = app.Command("download-artifact", "Download, unpack and cache artifact from GitHub.")
	downloadArtifactCommandName     = downloadArtifactCommand.Flag("name", "The artifact name.").Short('n').Required().String()
	downloadArtifactCommandUrl      = downloadArtifactCommand.Flag("url", "The artifact URL.").Short('u').Required().String()
	downloadArtifactCommandChecksum = downloadArtifactCommand.Flag("sha512", "The expected sha512 of file.").String()
)

func main() {
	// otherwise error: duplicate long flag --version, try --help
	// kingpin cannot correctly process it
	app.VersionFlag = nil

	log.SetHandler(log_cli.Default)

	debugEnv, isDebugDefined := os.LookupEnv("DEBUG")
	if isDebugDefined && debugEnv != "false" {
		log.SetLevel(log.DebugLevel)
	}

	if os.Getenv("SZA_ARCHIVE_TYPE") != "" {
		err := compress()
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		return
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
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

	case downloadCommand.FullCommand():
		err := download.Download(*downloadCommandUrl, *downloadCommandOutput, *downloadCommandChecksum)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case downloadArtifactCommand.FullCommand():
		dirPath, err := download.DownloadArtifact(*downloadArtifactCommandName, *downloadArtifactCommandUrl, *downloadArtifactCommandChecksum)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		_, err = os.Stdout.Write([]byte(dirPath))
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case copyDirCommand.FullCommand():
		err := util.CopyDirOrFile(*copyDirSource, *copyDirDestination)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
}

func cleanUpSnap(dir string) (error) {
	unnecessaryFiles := []string{
		"usr/share/doc",
		"usr/share/man",
		"usr/share/icons",
		"usr/share/bash-completion",
		"usr/share/lintian",
		"usr/share/dh-python",
		"usr/share/python3",

		"usr/lib/python*",
		"usr/bin/python*",
	}

	sem := make(chan bool, 4)
	for _, file := range unnecessaryFiles {
		sem <- true
		go func() {
			defer func() { <-sem }()
			err := util.RemoveByGlob(filepath.Join(dir, file))
			log.Fatalf("%+v\n", errors.WithStack(err))
			if err != nil {
				log.Fatalf("%+v\n", errors.WithStack(err))
			}
		}()
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
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
