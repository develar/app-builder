package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"github.com/develar/app-builder/commands"
	"log"
)

var (
	app = kingpin.New("app-builder", "app-builder").Version("0.1.0")

	icnsToPng        = app.Command("icns-to-png", "convert ICNS to PNG")
	icnsToPngInFile  = icnsToPng.Flag("input", "input ICNS file").Short('i').Required().String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case icnsToPng.FullCommand():
		err := commands.ConvertIcnsToPng(*icnsToPngInFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}
