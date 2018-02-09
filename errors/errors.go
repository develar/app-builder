package errors

//noinspection GoSnakeCaseUsage
import (
	"github.com/apex/log"
	_errors "github.com/pkg/errors"
)

type causer interface {
	Cause() error
}

func LogErrorAndExit(err error) {
	var lastErrorWithCause error
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		lastErrorWithCause = err
		err = cause.Cause()
	}

	log.Fatalf("%+v\n", lastErrorWithCause)
}

func WithStack(err error) error {
	_, ok := err.(causer)
	if ok {
		return err
	} else {
		return _errors.WithStack(err)
	}
}

func Cause(err error) error {
	return _errors.Cause(err)
}

func Errorf(format string, args ...interface{}) error {
	return _errors.Errorf(format, args)
}