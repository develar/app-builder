package icons

import "fmt"

type ImageSizeError struct {
	File            string
	RequiredMinSize int
	errorCode       string
}

type ImageFormatError struct {
	File      string
	errorCode string
}

func (e *ImageSizeError) ErrorCode() string {
	return e.errorCode
}

func (e *ImageFormatError) ErrorCode() string {
	return e.errorCode
}

func (e *ImageSizeError) Error() string {
	return fmt.Sprintf("image %s must be at least %dx%d", e.File, e.RequiredMinSize, e.RequiredMinSize)
}

func (e *ImageFormatError) Error() string {
	return fmt.Sprintf("image %s shas unknown format", e.File)
}

func NewImageSizeError(file string, requiredMinSize int) *ImageSizeError {
	return &ImageSizeError{file, requiredMinSize, "ERR_ICON_TOO_SMALL"}
}
