package logger

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(LevelInfo, WithWriter(&buf))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	l.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected output to contain 'hello', got %q", buf.String())
	}
}

func TestNewWithFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "sub", "test.log")

	l, err := New(LevelInfo, WithFile(logPath))
	if err != nil {
		t.Fatalf("New() with file error: %v", err)
	}
	l.Info("file test")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !strings.Contains(string(data), "file test") {
		t.Errorf("expected log file to contain 'file test', got %q", string(data))
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"Info", LevelInfo},
		{"event", LevelEvent},
		{"EVENT", LevelEvent},
		{"unknown", LevelDisabled},
		{"", LevelDisabled},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestScope(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(LevelInfo, WithWriter(&buf))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	scoped := l.Scope("db")
	scoped.Info("query executed")
	if !strings.Contains(buf.String(), "db") {
		t.Errorf("expected scoped output to contain 'db', got %q", buf.String())
	}
}

func TestAllMethods(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(LevelDebug, WithWriter(&buf))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	l.Info("info msg")
	l.Event("event msg")
	l.Debug("key", "value")
	l.Error(errors.New("test error"), "error msg")

	out := buf.String()
	for _, want := range []string{"info msg", "event msg", "key", "value", "error msg"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(LevelInfo, WithWriter(&buf))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	l.Debug("should-not-appear", "value")
	l.Info("should-appear")

	out := buf.String()
	if strings.Contains(out, "should-not-appear") {
		t.Errorf("debug message should be filtered at info level, got %q", out)
	}
	if !strings.Contains(out, "should-appear") {
		t.Errorf("info message should appear at info level, got %q", out)
	}
}

func TestDisabledLevel(t *testing.T) {
	var buf bytes.Buffer
	l, err := New(LevelDisabled, WithWriter(&buf))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	l.Info("silent")
	l.Error(errors.New("err"), "silent error")

	if buf.Len() != 0 {
		t.Errorf("expected no output at disabled level, got %q", buf.String())
	}
}

func TestNewNoop(t *testing.T) {
	var l Logger = NewNoop()

	// Should not panic.
	l.Info("test")
	l.Event("test")
	l.Debug("k", "v")
	l.Error(errors.New("err"), "test")
	l.Scope("sub").Info("test")
}
