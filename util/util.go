package util

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/apex/log"
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

func WriteJsonToStdOut(v interface{}) error {
	serializedInputInfo, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(serializedInputInfo)
	if err != nil {
		return err
	}

	return nil
}

func Execute(command *exec.Cmd, currentWorkingDirectory string) error {
	log.WithFields(log.Fields{
		"path": command.Path,
		"args": command.Args,
	}).Debug("execute command")

	if currentWorkingDirectory != "" {
		command.Dir = currentWorkingDirectory
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return errors.WithStack(command.Run())
}
