package reporter

import (
	"fmt"
	"os"
	"strings"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

func init() {
	initLogger()
}

func initLogger() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	rl := rz.InfoLevel

	switch strings.ToLower(level) {
	case "debug":
		rl = rz.DebugLevel
	case "info":
		rl = rz.InfoLevel
	case "warn":
		rl = rz.WarnLevel
	case "error":
		rl = rz.ErrorLevel
	default:
		fmt.Fprintf(os.Stderr, "unknown log level `%s`", level)
	}

	log.SetLogger(log.With(
		rz.Level(rl),
		rz.Formatter(rz.FormatterConsole()),
	))
}
