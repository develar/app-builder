package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/apex/log"
	"github.com/develar/errors"
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
	logCommandExecuting(command)

	if currentWorkingDirectory != "" {
		command.Dir = currentWorkingDirectory
	}

	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		return errors.WithMessage(err, "output: " + output.String())
	}

	return nil
}

func logCommandExecuting(command *exec.Cmd) {
	log.WithFields(log.Fields{
		"path": command.Path,
		"args": command.Args,
	}).Debug("execute command")
}

func LogErrorAndExit(err error) {
	log.Fatalf("%+v\n", err)
}