package icons

import (
	"image"
	"os"

	"github.com/pkg/errors"
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
