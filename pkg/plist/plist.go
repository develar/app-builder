package plist

import (
	"bytes"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"howett.net/plist"
)

func ConfigurePlistCommand(app *kingpin.Application) {
	command := app.Command("decode-plist", "")
	files := command.Flag("file", "").Short('f').Required().Strings()
	command.Action(func(context *kingpin.ParseContext) error {
		return decode(*files)
	})
}

func decode(files []string) error {
	results := make([][]byte, len(files))
	err := util.MapAsync(len(files), func(index int) (func() error, error) {
		filePath := files[index]
		return func() error {
			file, err := os.Open(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					results[index] = nil
					return nil
				}
				return errors.WithStack(err)
			}

			defer util.Close(file)
			decoder := plist.NewDecoder(file)
			value := make(map[string]interface{})
			err = decoder.Decode(&value)
			if err != nil {
				return errors.WithStack(err)
			}

			jsonData, err := jsoniter.Marshal(value)
			if err != nil {
				return errors.WithStack(err)
			}

			results[index] = jsonData

			return nil
		}, nil
	})
	var b bytes.Buffer
	b.WriteString("[")
	for index, value := range results {
		if index != 0 {
			b.WriteString(",")
		}

		if len(value) == 0 {
			b.WriteString("null")
		} else {
			b.Write(value)
		}
	}
	b.WriteString("]")
	_, _ = os.Stdout.Write(b.Bytes())
	return errors.WithStack(err)
}
