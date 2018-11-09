package util

import (
	"io"
	"os"
	"os/exec"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
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
	serializedInputInfo, err := jsoniter.ConfigFastest.Marshal(v)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(serializedInputInfo)
	_ = os.Stdout.Close()
	return err
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
		return errors.Errorf("error: %s\npath: %s\nargs: %s\noutput: %s", err, command.Path, command.Args, output)
	} else if IsDebugEnabled() && len(output) != 0 {
		log.Debug(string(output))
	}

	return nil
}

func StartPipedCommands(producer *exec.Cmd, consumer *exec.Cmd) error {
	err := producer.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	err = consumer.Start()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func RunPipedCommands(producer *exec.Cmd, consumer *exec.Cmd) error {
	err := StartPipedCommands(producer, consumer)
	if err != nil {
		return errors.WithStack(err)
	}

	err = WaitPipedCommand(producer, consumer)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func WaitPipedCommand(producer *exec.Cmd, consumer *exec.Cmd) error {
	err := producer.Wait()
	if err != nil {
		return errors.WithStack(err)
	}

	err = consumer.Wait()
	if err != nil {
		return errors.WithStack(err)
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

// http://www.blevesearch.com/news/Deferred-Cleanup,-Checking-Errors,-and-Potential-Problems/
func Close(c io.Closer) {
	err := c.Close()
	if err != nil && err != os.ErrClosed {
		if e, ok := err.(*os.PathError); ok && e.Err == os.ErrClosed {
			return
		}
		log.Errorf("%v", err)
	}
}