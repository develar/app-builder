package util

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/errors"
)

func ConfigureIsRemoveStageParam(command *kingpin.CmdClause) *bool {
	var isRemoveStageDefaultValue string
	if IsDebugEnabled() && !IsEnvTrue("BUILDER_REMOVE_STAGE_EVEN_IF_DEBUG") {
		isRemoveStageDefaultValue = "false"
	} else {
		isRemoveStageDefaultValue = "true"
	}

	return command.Flag("remove-stage", "Whether to remove stage after build.").Default(isRemoveStageDefaultValue).Bool()
}

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

func IsDebugEnabled() bool {
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

// useful for snap, where prime command took a lot of time and we need to read progress messages
func ExecuteWithInheritedStdOutAndStdErr(command *exec.Cmd, currentWorkingDirectory string) error {
	preCommandExecute(command, currentWorkingDirectory)

	// not an error - command error output printed to out stdout (like logging)
	command.Stdout = os.Stderr
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func Execute(command *exec.Cmd, currentWorkingDirectory string) error {
	preCommandExecute(command, currentWorkingDirectory)

	output, err := command.CombinedOutput()
	if err != nil {
		return errors.WithMessage(err, "output: "+string(output))
	} else if IsDebugEnabled() && len(output) != 0 {
		log.Debug(string(output))
	}

	return nil
}

func preCommandExecute(command *exec.Cmd, currentWorkingDirectory string) {
	if currentWorkingDirectory != "" {
		command.Dir = currentWorkingDirectory
	}

	log.WithFields(log.Fields{
		"path": command.Path,
		"args": command.Args,
	}).Debug("execute command")
}

func LogErrorAndExit(err error) {
	log.Fatalf("%+v\n", err)
}