package util

import (
	"os"
	"os/exec"

	"github.com/develar/errors"
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
}

func (e *ExecError) Error() string {
	return e.Cause.Error()
}