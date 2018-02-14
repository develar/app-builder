package icons

import (
	"bufio"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/biessek/golang-ico"
	"github.com/develar/errors"
)

const (
	PNG = 0
	ICO = 1
)

func DecodeImageConfig(file string) (*image.Config, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result, _, err := image.DecodeConfig(reader)
	if err != nil {
		reader.Close()

		if err == image.ErrFormat {
			err = &ImageFormatError{file, "ERR_ICON_UNKNOWN_FORMAT"}
		}
		return nil, errors.WithStack(err)
	}

	err = reader.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func DecodeImageAndClose(reader io.Reader, closer io.Closer) (image.Image, error) {
	result, _, err := image.Decode(reader)
	if err != nil {
		closer.Close()
		return nil, errors.WithStack(err)
	}

	err = closer.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

func SaveImage(image image.Image, outFileName string) error {
	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}

	return SaveImage2(image, outFile, PNG)
}

func SaveImage2(image image.Image, outFile *os.File, format int) error {
	writer := bufio.NewWriter(outFile)

	var err error
	if format == PNG {
		err = png.Encode(writer, image)
	} else {
		err = ico.Encode(writer, image)
	}

	if err != nil {
		outFile.Close()
		return err
	}

	flushError := writer.Flush()
	closeError := outFile.Close()
	if flushError != nil {
		return flushError
	}
	if closeError != nil {
		return closeError
	}

	return nil
}