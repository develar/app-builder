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
	_, _ = fmt.Fprintf(h.Writer, "%s%s", e.Message, strings.Repeat(" ", max(1, 15 /* because first field adds space before */ - len(e.Message))))

	for _, name := range names {
		if name == "source" {
			continue
		}
		_, _ = fmt.Fprintf(h.Writer, " %s=%v", myColor.Sprint(name), e.Fields.Get(name))
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