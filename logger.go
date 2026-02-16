// Package logger provides a structured logging interface built on top of zerolog.
//
// It supports scoped loggers, colored console output, file logging, and a custom
// Event level between Info and Warn for domain-significant occurrences.
//
// Basic usage:
//
//	l, err := logger.New(logger.LevelInfo)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	l.Info("server started")
//	l.Scope("auth").Info("user logged in")
//
// With file logging:
//
//	l, err := logger.New(logger.LevelDebug, logger.WithFile("logs/app.log"))
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

// Level controls the minimum severity of messages that are logged.
type Level int

const (
	LevelDebug    Level = iota // LevelDebug logs everything.
	LevelInfo                  // LevelInfo logs info, events, warnings, and errors.
	LevelEvent                 // LevelEvent logs events, warnings, and errors.
	LevelDisabled              // LevelDisabled silences all output.
)

// eventLevel is a custom zerolog level for domain events (between Info and Warn).
const eventLevel = zerolog.Level(2)

var levelMap = map[Level]zerolog.Level{
	LevelDebug:    zerolog.DebugLevel,
	LevelInfo:     zerolog.InfoLevel,
	LevelEvent:    eventLevel,
	LevelDisabled: zerolog.Disabled,
}

// ParseLevel converts a string to a Level. It is case-insensitive.
// Unrecognized strings return LevelDisabled.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "event":
		return LevelEvent
	default:
		return LevelDisabled
	}
}

// Logger is the interface for structured logging with scoping support.
type Logger interface {
	Info(string)
	Event(string)
	Debug(string, string)
	Error(error, string)
	Scope(string) Logger
}

// Option configures a Logger created by New.
type Option func(*options)

type options struct {
	filePath string
	writer   io.Writer
}

// WithFile enables additional file logging at the given path.
// Parent directories are created automatically if they don't exist.
func WithFile(path string) Option {
	return func(o *options) {
		o.filePath = path
	}
}

// WithWriter sets a custom writer for console output instead of os.Stdout.
// This is useful for testing or redirecting output.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

// New creates a Logger at the given level. By default it writes colored output
// to stdout. Use WithFile or WithWriter to customize output destinations.
func New(level Level, opts ...Option) (Logger, error) {
	cfg := &options{
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	writers := []io.Writer{newConsoleWriter(cfg.writer)}

	if cfg.filePath != "" {
		dir := filepath.Dir(cfg.filePath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("logger: create log directory: %w", err)
		}
		f, err := os.OpenFile(cfg.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("logger: open log file: %w", err)
		}
		writers = append(writers, f)
	}

	zlevel, ok := levelMap[level]
	if !ok {
		zlevel = zerolog.Disabled
	}

	zl := zerolog.
		New(zerolog.MultiLevelWriter(writers...)).
		Level(zlevel).
		With().
		Timestamp().
		Logger()

	return &logger{zl, "main"}, nil
}

func newConsoleWriter(w io.Writer) zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: zerolog.TimeFormatUnix,
		FormatLevel: func(i interface{}) string {
			const (
				colorReset  = "\x1b[0m"
				colorRed    = "\x1b[31m"
				colorGreen  = "\x1b[32m"
				colorYellow = "\x1b[33m"
				colorCyan   = "\x1b[36m"
				colorWhite  = "\x1b[37m"
			)
			if s, ok := i.(string); ok {
				switch s {
				case "event":
					return colorWhite + "EVT" + colorReset
				case "info":
					return colorGreen + "INF" + colorReset
				case "debug":
					return colorCyan + "DBG" + colorReset
				case "error":
					return colorRed + "ERR" + colorReset
				case "warn":
					return colorYellow + "WRN" + colorReset
				default:
					return strings.ToUpper(s)
				}
			}
			return "???"
		},
	}
}

type logger struct {
	zl    zerolog.Logger
	scope string
}

func init() {
	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		if l == eventLevel {
			return "event"
		}
		return l.String()
	}
}

func (l *logger) Info(msg string) {
	l.zl.Info().Msgf("[%s] %s", l.scope, msg)
}

func (l *logger) Event(msg string) {
	l.zl.WithLevel(eventLevel).Msgf("-%s (%s)", msg, l.scope)
}

func (l *logger) Debug(key, value string) {
	l.zl.Debug().Msgf(" %s: %s (%s)", key, value, l.scope)
}

func (l *logger) Error(err error, msg string) {
	l.zl.Error().Err(err).Msgf("[%s] %s", l.scope, msg)
}

func (l *logger) Scope(name string) Logger {
	return &logger{l.zl.With().Logger(), name}
}

