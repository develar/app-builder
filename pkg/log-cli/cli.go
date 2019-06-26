// Package cli implements a colored text handler suitable for command-line interfaces.
package log_cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

func InitLogger() {
	log.SetHandler(Default)
	debugEnv, isDebugDefined := os.LookupEnv("DEBUG")
	if isDebugDefined && debugEnv != "false" {
		log.SetLevel(log.DebugLevel)
	}

	forceColor := os.Getenv("FORCE_COLOR")
	if forceColor == "0" {
		color.NoColor = true
	} else if forceColor == "1" || forceColor == "true" {
		color.NoColor = false
	}
}

// Default handler outputting to stderr.
var Default = New(os.Stderr)

// Colors mapping.
var Colors = [...]*color.Color{
	log.DebugLevel: color.New(color.FgWhite),
	log.InfoLevel:  color.New(color.FgBlue),
	log.WarnLevel:  color.New(color.FgYellow),
	log.ErrorLevel: color.New(color.FgRed),
	log.FatalLevel: color.New(color.FgRed),
}

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "•",
	log.InfoLevel:  "•",
	log.WarnLevel:  "•",
	log.ErrorLevel: "⨯",
	log.FatalLevel: "⨯",
}

// Handler implementation.
type Handler struct {
	mu      sync.Mutex
	Writer  io.Writer
	Padding int
}

// New handler.
func New(w io.Writer) *Handler {
	if !color.NoColor {
		if f, ok := w.(*os.File); ok {
			return &Handler{
				Writer:  colorable.NewColorable(f),
				Padding: 2,
			}
		}
	}

	return &Handler{
		Writer:  w,
		Padding: 2,
	}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(e *log.Entry) error {
	myColor := Colors[e.Level]
	level := Strings[e.Level]
	names := e.Fields.Names()

	h.mu.Lock()
	defer h.mu.Unlock()

	_, _ = myColor.Fprintf(h.Writer, "%*s ", h.Padding+1, level)
	fieldOffset := h.Padding + 1 + 1

	if e.Level >= log.ErrorLevel {
		_, _ = myColor.Fprint(h.Writer, e.Message)
	} else {
		_, _ = h.Writer.Write([]byte(e.Message))
	}

	fieldOffset += len(e.Message)

	n, _ := h.Writer.Write([]byte(strings.Repeat(" ", max(2, 16-len(e.Message)))))
	// n can be used because ASCII only (so, byte count equals to char count)
	fieldOffset += n

	fieldNameAndValueList := make([]string, len(names))

	totalLength := 0
	for index, name := range names {
		if name == "source" {
			continue
		}

		v := fmt.Sprintf("%s=%v", myColor.Sprint(name), e.Fields.Get(name))
		fieldNameAndValueList[index] = v
		totalLength += len(v)
	}

	stacked := totalLength > 160
	var fieldPrefix []byte
	if stacked {
		fieldPrefix = make([]byte, fieldOffset + 1)
		fieldPrefix[0] = 10
		for i := 1; i < len(fieldPrefix); i++ {
			fieldPrefix[i] = 32
		}
	} else {
		fieldPrefix = []byte(" ")
	}

	writtenIndex := 0
	for _, v := range fieldNameAndValueList {
		if len(v) == 0 {
			continue
		}

		if writtenIndex > 0 {
			_, _ = h.Writer.Write(fieldPrefix)
		}

		_, _ = h.Writer.Write([]byte(v))

		writtenIndex++
	}

	_, _ = fmt.Fprintln(h.Writer)

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}