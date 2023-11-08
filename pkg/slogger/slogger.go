package slogger

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-cz/devslog"
)

type ProductionHandler struct {
	slog.Handler
}

func New(appName string, version string, production bool) (*slog.Logger, error) {
	if appName == "" {
		return nil, errors.New("appName is not defined")
	}

	level := slog.LevelDebug
	if production {
		level = slog.LevelInfo
	}

	handlerOptions := &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: getReplaceAttr(production),
	}

	defaultAttrs := []slog.Attr{
		slog.String("app", appName),
		slog.String("release", version),
	}

	handler := setDefaultHandler(handlerOptions, defaultAttrs, production)

	return handler, nil
}

func setDefaultHandler(handlerOptions *slog.HandlerOptions, attrs []slog.Attr, production bool) *slog.Logger {
	if production {
		return slog.New(&ProductionHandler{
			Handler: slog.NewJSONHandler(os.Stdout, handlerOptions).WithAttrs(attrs),
		})
	} else {
		opts := &devslog.Options{
			HandlerOptions:     handlerOptions,
			MaxSlicePrintSize:  20,
			SortKeys:           true,
			NewLineAfterLog:    true,
			MaxErrorStackTrace: 2,
		}
		return slog.New(devslog.NewHandler(os.Stdout, opts))
	}
}

const (
	LevelTrace = slog.Level(-8)
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

func getReplaceAttr(production bool) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		// Converting time of log to UTC format
		if a.Key == slog.TimeKey {
			inputLayout := "2006-01-02 15:04:05.999999999 -0700 MST"
			outputLayout := "2006-01-02T15:04:05Z0700"

			inputTime, err := time.Parse(inputLayout, a.Value.String())
			if err != nil {
				panic(err)
			}

			outputTime := inputTime.UTC()
			a.Value = slog.StringValue(outputTime.Format(outputLayout))
		}

		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)

			switch {
			case level < LevelDebug:
				a.Value = slog.StringValue("TRACE")
			}
		}

		// Converting log level to lowercase
		if production {
			a.Value = slog.StringValue(strings.ToLower(a.Value.String()))
		}

		// Changing key from "source" to "caller"
		// For now it is commented, if you want to use key name as "caller", just uncomment next 4 lines
		// if a.Key == slog.SourceKey {
		// 	a.Key = "caller"
		// 	return a
		// }

		return a
	}
}

// errorCause recursively unwraps given error and returns the topmost
// non-nil error cause, same as github.com/pkg/errors.Cause(err).
func ErrorCause(err error) (cause error) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		cause = e
	}
	return cause
}

func ErrorCauseString(err error, prefixes ...string) (cause string) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		err = e
	}

	for _, p := range prefixes {
		cause += p + ": "
	}

	return cause + err.Error()
}
