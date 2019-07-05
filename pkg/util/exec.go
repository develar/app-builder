package util

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/errors"
	"go.uber.org/zap"
)

// useful for snap, where prime command took a lot of time and we need to read progress messages
func ExecuteAndPipeStdOutAndStdErr(command *exec.Cmd) error {
	preCommandExecute(command)

	// not an error - command error output printed to out stdout (like logging)
	command.Stdout = os.Stderr
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type ExecError struct {
	Cause            error
	CommandAndArgs   []string
	WorkingDirectory string

	Output      []byte
	ErrorOutput []byte

	Message     string
	ExtraFields []zap.Field
}

func (e *ExecError) Error() string {
	return e.Cause.Error()
}

func Execute(command *exec.Cmd) ([]byte, error) {
	preCommandExecute(command)

	var output bytes.Buffer
	command.Stdout = &output

	var errorOutput bytes.Buffer
	command.Stderr = &errorOutput

	err := command.Run()
	if err != nil {
		return output.Bytes(), &ExecError{
			Cause:            err,
			CommandAndArgs:   command.Args,
			WorkingDirectory: command.Dir,

			Output:      output.Bytes(),
			ErrorOutput: errorOutput.Bytes(),
		}
	} else if log.IsDebugEnabled() && !(strings.HasSuffix(command.Path, "openssl") || strings.HasSuffix(command.Path, "openssl.exe")) {
		var fields []zap.Field
		fields = append(fields, zap.String("executable", command.Args[0]))
		if output.Len() > 0 {
			fields = append(fields, zap.String("out", output.String()))
		}
		if errorOutput.Len() > 0 {
			fields = append(fields, zap.String("errorOut", errorOutput.String()))
		}
		log.Debug("command executed", fields...)
	}

	return output.Bytes(), nil
}

func preCommandExecute(command *exec.Cmd) {
	if log.IsDebugEnabled() {
		log.Debug("execute command", zap.String("command", argListToSafeString(command.Args)), zap.String("workingDirectory", command.Dir))
	}
}
