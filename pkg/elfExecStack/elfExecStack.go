package elfExecStack

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/errors"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("clear-exec-stack", "")

	file := command.Flag("input", "").Short('i').Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		err := ClearExecStack(*file)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
}

func ClearExecStack(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	defer file.Close()

	data, err := ioutil.ReadAll(io.LimitReader(file, 2048))
	if err != nil {
		return errors.WithStack(err)
	}

	// GNU_STACK 1685382481
	gnuStack := []byte{0x51, 0xE5, 0x74, 0x64}
	index := bytes.Index(data, gnuStack)
	if index < 0 {
		return errors.Errorf("cannot find GNU_STACK header in a first 2048 bytes")
	}

	flagIndex := index + len(gnuStack)

	if index >= len(data) {
		return errors.Errorf("GNU_STACK header flags outside of first 2048 bytes")
	}

	flagValue := data[flagIndex]
	if flagValue == 6 {
		// already cleared
		log.WithField("flags", flagValue).Debug("stack is already cleared")
		return nil
	}

	_, err = file.WriteAt([]byte{0x06}, int64(flagIndex))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
