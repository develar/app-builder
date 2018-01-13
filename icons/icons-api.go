package icons

import (
	"bufio"
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

type IconConvertResult struct {
	File string `json:"file"`
}

type MisConfigurationError struct {
	Message string `json:"error"`
	Code    string `json:"errorCode"`
}

// sorted by suitability
var icnsTypesForIco = []string{ICNS_256, ICNS_256_RETINA, ICNS_512, ICNS_512_RETINA, ICNS_1024}

func LoadImage(file string) (image.Image, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer reader.Close()

	bufferedReader := bufio.NewReader(reader)

	icns, err := isIcns(bufferedReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if icns {
		subImageInfoList, err := ReadIcns(bufferedReader)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, osType := range icnsTypesForIco {
			subImage, ok := subImageInfoList[osType]
			if ok {
				reader.Seek(int64(subImage.Offset), 0)
				bufferedReader.Reset(reader)
				return DecodeImageAndClose(bufferedReader, reader)
			}
		}

		return nil, NewImageSizeError(file, 256)
	}

	return DecodeImageAndClose(bufferedReader, reader)
}

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
