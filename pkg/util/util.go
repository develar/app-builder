package util

import (
	"crypto/sha512"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strings"

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
		return errors.WithStack(err)
	}

	_, err = os.Stdout.Write(serializedInputInfo)
	_ = os.Stdout.Close()
	return errors.WithStack(err)
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

func Execute(command *exec.Cmd, currentWorkingDirectory string) ([]byte, error) {
	preCommandExecute(command, currentWorkingDirectory)

	output, err := command.Output()
	if err != nil {
		errorOut := ""
		if exitError, ok := err.(*exec.ExitError); ok {
			errorOut = string(exitError.Stderr)
		}

		return nil, errors.New("error: " + err.Error() +
			"\npath: " + command.Path +
			"\nargs: " + argListToSafeString(command.Args) +
			"\noutput: " + string(output) +
			"\nerror output:" + errorOut)
	} else if IsDebugEnabled() && len(output) != 0 && !(strings.HasSuffix(command.Path, "openssl") || strings.HasSuffix(command.Path, "openssl.exe")) {
		log.Debug(string(output))
	}

	return output, nil
}

func argListToSafeString(args []string) string {
	var result strings.Builder
	for index, value := range args {
		if strings.HasPrefix(value, "pass:") {
			hasher := sha512.New()
			_, err := hasher.Write([]byte(value))
			if err == nil {
				value = "sha512-first-8-chars-" + hex.EncodeToString(hasher.Sum(nil)[0:4])
			} else {
				log.WithError(err).Warn("cannot compute sha512 hash of password to log")
				value = "<hidden>"
			}
		}
		if index > 0 {
			result.WriteRune(' ')
		}
		result.WriteString(value)
	}

	return result.String()
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
		"args": argListToSafeString(command.Args),
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
