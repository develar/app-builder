package icons

import (
	"bufio"
	"image"
	"os"

	"github.com/develar/errors"
)

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

type IconConvertResult struct {
	Icons []IconInfo `json:"icons"`
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

	isIcns, err := IsIcns(bufferedReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if isIcns {
		subImageInfoList, err := ReadIcns(bufferedReader)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, osType := range icnsTypesForIco {
			subImage, ok := subImageInfoList[osType]
			if ok {
				reader.Seek(int64(subImage.Offset), 0)
				bufferedReader.Reset(reader)
				// golang doesn't support JPEG2000
				return DecodeImageAndClose(bufferedReader, reader)
			}
		}

		return nil, NewImageSizeError(file, 256)
	}

	return DecodeImageAndClose(bufferedReader, reader)
}

type InputFileInfo struct {
	MaxIconSize int
	MaxIconPath string
	SizeToPath  map[int]string

	maxImage image.Image

	recommendedMinSize int
}
