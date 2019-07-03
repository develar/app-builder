package log

import (
	"io"
	"os"

	"github.com/develar/app-builder/pkg/zap-cli-encoder"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LOG *zap.Logger

func InitLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	level := zapcore.InfoLevel
	debugEnv, isDebugDefined := os.LookupEnv("DEBUG")
	if isDebugDefined && debugEnv != "false" {
		level = zapcore.DebugLevel
	}

	colored := isColored()
	var writer io.Writer
	if colored {
		writer = colorable.NewColorableStderr()
	} else {
		writer = os.Stderr
	}
	LOG = zap.New(zapcore.NewCore(
		zap_cli_encoder.NewConsoleEncoder(encoderConfig, colored),
		zapcore.AddSync(writer),
		level,
	))
}

func isColored() bool {
	forceColor, ok := os.LookupEnv("FORCE_COLOR")
	if ok && (forceColor == "1" || forceColor == "true" || forceColor == "") {
		return true
	}

	if forceColor == "0" || forceColor == "false" || os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())) {
		return false
	}
	return true
}

func Warn(msg string, fields ...zapcore.Field) {
	LOG.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	LOG.Error(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	LOG.Info(msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	LOG.Debug(msg, fields...)
}

func IsDebugEnabled() bool {
	return LOG.Core().Enabled(zap.DebugLevel)
}