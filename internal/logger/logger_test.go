package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/autobrr/harbrr/internal/config"
	"github.com/autobrr/harbrr/internal/logger"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     config.LogConfig
		wantErr bool
	}{
		{"json info", config.LogConfig{Level: "info", Format: "json"}, false},
		{"console debug", config.LogConfig{Level: "debug", Format: "console"}, false},
		{"bad level", config.LogConfig{Level: "loud", Format: "json"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			lg, err := logger.New(tt.cfg, &buf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("New() = nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() = %v", err)
			}
			lg.Info().Msg("hello")
			if !strings.Contains(buf.String(), "hello") {
				t.Errorf("log output %q missing message", buf.String())
			}
		})
	}
}

func TestNewLevelFilters(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	lg, err := logger.New(config.LogConfig{Level: "warn", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	lg.Info().Msg("filtered")
	lg.Warn().Msg("kept")

	out := buf.String()
	if strings.Contains(out, "filtered") {
		t.Errorf("info message should be filtered at warn level: %q", out)
	}
	if !strings.Contains(out, "kept") {
		t.Errorf("warn message should pass: %q", out)
	}
}
