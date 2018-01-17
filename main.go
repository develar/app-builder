package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	logCli "github.com/apex/log/handlers/cli"
	"github.com/develar/app-builder/asar"
	"github.com/develar/app-builder/blockmap"
	"github.com/develar/app-builder/icons"
	"github.com/pkg/errors"
)

var (
	app = kingpin.New("app-builder", "app-builder").Version("0.6.0")

	convertIcon          = app.Command("icon", "create ICNS or ICO or icon set from PNG files")
	convertIconSources   = convertIcon.Flag("input", "input directory or file").Short('i').Required().Strings()
	convertIconOutFormat = convertIcon.Flag("format", "output format").Short('f').Required().Enum("icns", "ico", "set")
	convertIconRoots     = convertIcon.Flag("root", "base directory to resolve relative path").Short('r').Strings()

	buildBlockmap            = app.Command("blockmap", "Generates file block map for differential update using content defined chunking (that is robust to insertions, deletions, and changes to input file)")
	buildBlockmapInFile      = buildBlockmap.Flag("input", "input file").Short('i').Required().String()
	buildBlockmapOutFile     = buildBlockmap.Flag("output", "output file").Short('o').String()
	buildBlockmapCompression = buildBlockmap.Flag("compression", "compression, one of: gzip, deflate").Short('c').Default("gzip").Enum("gzip", "deflate")

	buildAsar        = app.Command("asar", "")
	buildAsarOutFile = buildAsar.Flag("output", "").Required().String()
)

func main() {
	log.SetHandler(logCli.Default)

	debugEnv, isDebugDefined := os.LookupEnv("DEBUG")
	if isDebugDefined && debugEnv != "false" {
		log.SetLevel(log.DebugLevel)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case convertIcon.FullCommand():
		doConvertIcon()

	case buildAsar.FullCommand():
		err := asar.BuildAsar(*buildAsarOutFile)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case buildBlockmap.FullCommand():
		err := doBuildBlockMap()
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
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
	switch *buildBlockmapCompression {
	case "gzip":
		compressionFormat = blockmap.GZIP
	case "deflate":
		compressionFormat = blockmap.DEFLATE
	default:
		return fmt.Errorf("unknown compression format %s", *buildBlockmapCompression)
	}

	inputInfo, err := blockmap.BuildBlockMap(*buildBlockmapInFile, blockmap.DefaultChunkerConfiguration, compressionFormat, *buildBlockmapOutFile)
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
