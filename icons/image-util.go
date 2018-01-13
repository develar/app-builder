package icons

import (
	"image"
	"io"

	"github.com/pkg/errors"
)

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
