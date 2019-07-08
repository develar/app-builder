package util

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
	"gopkg.in/alessio/shellescape.v1"
)

func ConfigureIsRemoveStageParam(command *kingpin.CmdClause) *bool {
	var isRemoveStageDefaultValue string
	if log.IsDebugEnabled() && !IsEnvTrue("BUILDER_REMOVE_STAGE_EVEN_IF_DEBUG") {
		isRemoveStageDefaultValue = "false"
	} else {
		isRemoveStageDefaultValue = "true"
	}

	return command.Flag("remove-stage", "Whether to remove stage after build.").Default(isRemoveStageDefaultValue).Bool()
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

func argListToSafeString(args []string) string {
	var result strings.Builder
	for index, value := range args {
		if strings.HasPrefix(value, "pass:") {
			hasher := sha512.New()
			_, err := hasher.Write([]byte(value))
			if err == nil {
				value = "sha512-first-8-chars-" + hex.EncodeToString(hasher.Sum(nil)[0:4])
			} else {
				log.Warn("cannot compute sha512 hash of password to log", zap.Error(err))
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

func LogErrorAndExit(err error) {
	if execError, ok := err.(*ExecError); ok {
		message := execError.Message
		if len(message) == 0 {
			message = "cannot execute"
		}

		fields := execError.ExtraFields
		fields = append(fields, CreateExecErrorLogEntry(execError)...)
		log.LOG.Error(message, fields...)
		_ = log.LOG.Sync()
		// electron-builder in this case doesn't report app-builder error
		os.Exit(2)
	} else {
		log.LOG.Fatal(fmt.Sprintf("%+v", err))
	}
}

func CreateExecErrorLogEntry(execError *ExecError) []zap.Field {
	var fields []zap.Field
	fields = append(fields, zap.NamedError("cause", execError.Cause))
	if len(execError.Output) > 0 {
		fields = append(fields, zap.ByteString("out", execError.Output))
	}
	if len(execError.ErrorOutput) > 0 {
		fields = append(fields, zap.ByteString("errorOut", execError.ErrorOutput))
	}
	fields = append(fields,
		zap.String("command", argListToSafeString(execError.CommandAndArgs)),
		zap.String("workingDir", execError.WorkingDirectory),
	)
	return fields
}

// http://www.blevesearch.com/news/Deferred-Cleanup,-Checking-Errors,-and-Potential-Problems/
func Close(c io.Closer) {
	err := c.Close()
	if err != nil && err != os.ErrClosed && err != io.ErrClosedPipe {
		if e, ok := err.(*os.PathError); ok && e.Err == os.ErrClosed {
			return
		}
		log.Error("cannot close", zap.Error(err))
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