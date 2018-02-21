package icons

import (
	"image"
	"sort"
)

type IconInfo struct {
	File string `json:"file"`
	Size int    `json:"size"`
}

func sortBySize(list []IconInfo) {
	sort.Slice(list, func(i, j int) bool { return list[i].Size < list[j].Size })
	return
}

type IconConvertResult struct {
	Icons []IconInfo `json:"icons"`
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
