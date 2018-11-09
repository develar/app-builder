package blockmap

import (
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/util"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("blockmap", "Generates file block map for differential update using content defined chunking (that is robust to insertions, deletions, and changes to input file)")
	inFile := command.Flag("input", "input file").Short('i').Required().String()
	outFile := command.Flag("output", "output file").Short('o').String()
	compression := command.Flag("compression", "compression, one of: gzip, deflate").Short('c').Default("gzip").Enum("gzip", "deflate")

	command.Action(func(context *kingpin.ParseContext) error {
		var compressionFormat CompressionFormat
		switch *compression {
		case "gzip":
			compressionFormat = GZIP
		case "deflate":
			compressionFormat = DEFLATE
		default:
			return fmt.Errorf("unknown compression format %s", *compression)
		}

		inputInfo, err := BuildBlockMap(*inFile, DefaultChunkerConfiguration, compressionFormat, *outFile)
		if err != nil {
			return err
		}
		return util.WriteJsonToStdOut(inputInfo)
	})
}
