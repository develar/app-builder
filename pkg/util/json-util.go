package util

import (
	"os"

	"github.com/develar/errors"
	"github.com/json-iterator/go"
)

func WriteStringProperty(name string, value string, jsonWriter *jsoniter.Stream) {
	jsonWriter.WriteObjectField(name)
	jsonWriter.WriteString(value)
}

func FlushJsonWriterAndCloseOut(jsonWriter *jsoniter.Stream) error {
	err := jsonWriter.Flush()
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(os.Stdout.Close())
}
