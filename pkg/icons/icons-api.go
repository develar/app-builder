package icons

import (
	"image"
	"sort"

	"github.com/develar/errors"
)

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

func sortBySize(list []IconInfo) {
	sort.Slice(list, func(i, j int) bool { return list[i].Size < list[j].Size })
}

type IconConvertResult struct {
	Icons []IconInfo `json:"icons"`
	Fallback bool `json:"fallback"`
}

type MisConfigurationError struct {
	Message string `json:"error"`
	Code    string `json:"errorCode"`
}

type InputFileInfo struct {
	MaxIconSize int
	MaxIconPath string
	SizeToPath  map[int]string

	maxImage image.Image

	recommendedMinSize int
}

func (t *InputFileInfo) GetMaxImage() (image.Image, error) {
	if t.maxImage == nil {
		var err error
		t.maxImage, err = loadImage(t.MaxIconPath, t.recommendedMinSize)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return t.maxImage, nil
}
