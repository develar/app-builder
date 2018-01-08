package icons

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/biessek/golang-ico"
	"github.com/pkg/errors"
)

const (
	PNG = 0
	ICO = 1
)

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

type IconListResult struct {
	MaxIconPath string     `json:"maxIconPath"`
	MaxIconSize int        `json:"maxIconSize"`
	Icons       []IconInfo `json:"icons"`
}

type ImageSizeError struct {
	File            string
	RequiredMinSize int
	ErrorCode       string
}

func (e *ImageSizeError) Error() string {
	return fmt.Sprintf("image %s must be at least %dx%d", e.File, e.RequiredMinSize, e.RequiredMinSize)
}

func NewImageSizeError(file string, requiredMinSize int) *ImageSizeError {
	return &ImageSizeError{file, requiredMinSize, "ERR_TOO_SMALL_ICON"}
}

func LoadImage(file string) (image.Image, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result, _, err := image.Decode(reader)
	if err != nil {
		reader.Close()
		return nil, errors.WithStack(err)
	}

	err = reader.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

func DecodeImageConfig(file string) (*image.Config, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result, _, err := image.DecodeConfig(reader)
	if err != nil {
		reader.Close()
		return nil, errors.WithStack(err)
	}

	err = reader.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
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
