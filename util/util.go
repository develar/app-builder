package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"time"

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

func isDebugEnabled() bool {
	return getLevel() <= log.DebugLevel
}

func getLevel() log.Level {
	if logger, ok := log.Log.(*log.Logger); ok {
		return logger.Level
	}
	return log.InvalidLevel
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

func ExecuteWithTimeOut(command *exec.Cmd) (*bytes.Buffer, error) {
	logCommandExecuting(command)

	err := command.Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var output bytes.Buffer
	var errorOutput bytes.Buffer
	command.Stdout = &output
	command.Stderr = &errorOutput

	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()

	select {
	case <-time.After(30 * time.Second):
		err := command.Process.Kill()
		if err != nil {
			log.WithError(err).Error("failed to kill")
		}

		return nil, errors.Errorf("process killed as timeout reached")

	case err := <-done:
		if err != nil {
			log.Error(errorOutput.String())
			return &output, errors.WithStack(err)
		}

		if isDebugEnabled() {
			log.WithFields(log.Fields{
				"stdout": output.String(),
				"stderr": errorOutput.String(),
			}).Debug("output")
		}

		return &output, nil
	}
}

func Execute(command *exec.Cmd, currentWorkingDirectory string) error {
	logCommandExecuting(command)

	if currentWorkingDirectory != "" {
		command.Dir = currentWorkingDirectory
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return errors.WithStack(command.Run())
}

func logCommandExecuting(command *exec.Cmd) {
	log.WithFields(log.Fields{
		"path": command.Path,
		"args": command.Args,
	}).Debug("execute command")
}

func LogErrorAndExit(err error) {
	lastErrorWithCause := err
	for err != nil {
		cause, ok := err.(errors.Causer)
		if !ok {
			break
		}
		lastErrorWithCause = err
		err = cause.Cause()
	}

	log.Fatalf("%+v\n", lastErrorWithCause)
}