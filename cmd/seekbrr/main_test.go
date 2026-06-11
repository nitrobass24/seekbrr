package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/autobrr/seekbrr/internal/version"
)

// execute runs the command tree with args and returns combined stdout/stderr.
func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	out, err := execute(t, "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, version.String()) {
		t.Errorf("version output %q missing %q", out, version.String())
	}
}

func TestServeReportsNotImplemented(t *testing.T) {
	t.Parallel()
	out, err := execute(t, "serve")
	if err != nil {
		t.Fatalf("serve: %v", err)
	}
	if !strings.Contains(out, "not yet implemented") {
		t.Errorf("serve output %q missing not-implemented notice", out)
	}
	if !strings.Contains(out, "PLAINTEXT") {
		t.Errorf("serve output %q missing plaintext-secrets warning", out)
	}
}

func TestServeRejectsBadLogLevel(t *testing.T) {
	t.Parallel()
	if _, err := execute(t, "serve", "--log-level", "loud"); err == nil {
		t.Fatal("serve with invalid log level = nil error, want error")
	}
}
