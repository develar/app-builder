package util

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"gopkg.in/alessio/shellescape.v1"
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
	} else if IsDebugEnabled() && !(strings.HasSuffix(command.Path, "openssl") || strings.HasSuffix(command.Path, "openssl.exe")) {
		entry := log.WithField("command", command.Args[0])
		if output.Len() > 0 {
			entry = entry.WithField("out", output.String())
		}
		if errorOutput.Len() > 0 {
			entry = entry.WithField("errorOut", errorOutput.String())
		}
		entry.Debug("command executed")
	}

	return output.Bytes(), nil
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
		} else {
			value = shellescape.Quote(value)
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

func preCommandExecute(command *exec.Cmd) {
	log.WithFields(log.Fields{
		"args": argListToSafeString(command.Args),
		"workingDirectory": command.Dir,
	}).Debug("execute command")
}

func LogErrorAndExit(err error) {
	if execError, ok := err.(*ExecError); ok {
		CreateExecErrorLogEntry(execError).Fatal("cannot execute")
	} else {
		log.Fatalf("%+v\n", err)
	}
}

func CreateExecErrorLogEntry(execError *ExecError) *log.Entry {
	entry := log.WithField("cause", execError.Cause)
	if len(execError.Output) > 0 {
		entry = entry.WithField("out", string(execError.Output))
	}
	if len(execError.ErrorOutput) > 0 {
		entry = entry.WithField("errorOut", string(execError.ErrorOutput))
	}
	return entry.WithField("command", argListToSafeString(execError.CommandAndArgs)).WithField("workingDir", execError.WorkingDirectory).WithField("cause", execError.Cause)
}

// http://www.blevesearch.com/news/Deferred-Cleanup,-Checking-Errors,-and-Potential-Problems/
func Close(c io.Closer) {
	err := c.Close()
	if err != nil && err != os.ErrClosed && err != io.ErrClosedPipe {
		if e, ok := err.(*os.PathError); ok && e.Err == os.ErrClosed {
			return
		}
		log.Errorf("%v", err)
	}
}

func ContainsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func DecodeBase64IfNeeded(data string, v interface{}) error {
	if strings.HasPrefix(data, "{") || strings.HasPrefix(data, "[") {
		return jsoniter.UnmarshalFromString(data, v)
	} else {
		decodedData, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return errors.WithStack(err)
		}
		return jsoniter.Unmarshal(decodedData, v)
	}
}