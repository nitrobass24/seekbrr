// Package logger builds the zerolog logger from harbrr's typed config. It is
// intentionally small: level + output format only, growing as logging needs do.
package logger

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"

	"github.com/autobrr/harbrr/internal/config"
)

// New builds a zerolog.Logger from the log configuration, writing to w. A
// "console" format yields human-friendly colored output; "json" yields
// structured JSON. The level string is validated by config.Validate, but New
// re-parses defensively and wraps any error.
func New(cfg config.LogConfig, w io.Writer) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		return zerolog.Nop(), fmt.Errorf("logger: parse level %q: %w", cfg.Level, err)
	}

	out := w
	if cfg.Format == "console" {
		out = zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339}
	}

	return zerolog.New(out).Level(level).With().Timestamp().Logger(), nil
}
