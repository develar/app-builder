package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/icons"
)

var (
	app = kingpin.New("app-builder", "app-builder").Version("0.2.0")

	icnsToPng       = app.Command("icns-to-png", "convert ICNS to PNG")
	icnsToPngInFile = icnsToPng.Flag("input", "input ICNS file").Short('i').Required().String()

	pngToIcns          = app.Command("png-to-icns", "create ICNS from PNG files")
	pngToIcnsInFile = pngToIcns.Flag("input", "input directory or file").Short('i').Required().String()
	pngToIcnsRoots     = pngToIcns.Flag("root", "base directory to resolve relative path").Short('r').Required().Strings()

	collectIcons          = app.Command("collect-icons", "collect icons in a dir")
	collectIconsSourceDir = collectIcons.Flag("source", "source directory").Short('s').Required().String()
)

func main() {
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

	case pngToIcns.FullCommand():
		resultFile, err := icons.ConvertPngToIcns(*pngToIcnsInFile, *pngToIcnsRoots)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

		_, err = fmt.Printf("{\"file\":\"%s\"}", resultFile)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
}

func writeIconListResult(result *icons.IconListResult) error {
	serializedResult, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(serializedResult)
	return err
}
