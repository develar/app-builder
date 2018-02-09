package util

import (
	"io"

	"github.com/pkg/errors"
)

func CloseAndCheckError(err error, closable io.Closer) error {
	closeErr := closable.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	if closeErr != nil {
		return errors.WithStack(closeErr)
	}
	return nil
}