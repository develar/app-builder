package util

import (
	"os"

	"github.com/pkg/errors"
)

func ReadFile(file string, size int) ([]byte, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]byte, size)
	_, err = reader.Read(result)
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
