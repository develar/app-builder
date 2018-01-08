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
	app = kingpin.New("app-builder", "app-builder").Version("0.3.0")

	icnsToPng       = app.Command("icns-to-png", "convert ICNS to PNG")
	icnsToPngInFile = icnsToPng.Flag("input", "input ICNS file").Short('i').Required().String()

	convertIcon          = app.Command("icon", "create ICNS or ICO from PNG files")
	convertIconInFile    = convertIcon.Flag("input", "input directory or file").Short('i').Required().String()
	convertIconOutFormat = convertIcon.Flag("format", "output format").Short('f').Required().Enum("icns", "ico")
	convertIconRoots     = convertIcon.Flag("root", "base directory to resolve relative path").Short('r').Required().Strings()

	collectIcons          = app.Command("collect-icons", "collect icons in a dir")
	collectIconsSourceDir = collectIcons.Flag("source", "source directory").Short('s').Required().String()

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
	case icnsToPng.FullCommand():
		result, err := icons.ConvertIcnsToPng(*icnsToPngInFile)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		err = writeIconListResult(result)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

	case collectIcons.FullCommand():
		result, err := icons.CollectIcons(*collectIconsSourceDir)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		err = writeIconListResult(result)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

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
	resultFile, err := icons.ConvertIcon(*convertIconInFile, *convertIconRoots, *convertIconOutFormat)
	if err != nil {
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

	_, err = fmt.Printf("{\"file\":\"%s\"}", resultFile)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func printAppError(error icons.ImageError) {
	_, err := fmt.Printf("{\"error\":\"%s\", \"errorCode\": \"%s\"}", error, error.ErrorCode())
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

	serializedInputInfo, err := json.Marshal(inputInfo)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(serializedInputInfo)
	if err != nil {
		return err
	}

	return nil
}

func writeIconListResult(result *icons.IconListResult) error {
	serializedResult, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(serializedResult)
	return err
}
