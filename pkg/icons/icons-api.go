package icons

import (
	"image"
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

type InputFileInfo struct {
	MaxIconSize int
	MaxIconPath string
	SizeToPath  map[int]string

	maxImage image.Image

	recommendedMinSize int
}
