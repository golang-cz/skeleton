package slogger

import (
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/pkg/version"
)

type defaultHandler struct {
	slog.Handler
}

func Register(appName string, production bool) *slog.Logger {
	level := slog.LevelDebug
	if production {
		level = slog.LevelInfo
	}

	handlerOptions := slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replaceAttr,
	}

	defaultAttrs := []slog.Attr{
		slog.String("app", appName),
		slog.String("release", version.VERSION),
	}
	handler := setDefaultHandler(handlerOptions, defaultAttrs, production)
	slog.SetDefault(setDefaultHandler(handlerOptions, defaultAttrs, production))
	return handler
}

func setDefaultHandler(
	handlerOptions slog.HandlerOptions,
	attrs []slog.Attr,
	production bool,
) *slog.Logger {
	if production {
		return slog.New(&defaultHandler{
			Handler: handlerOptions.NewJSONHandler(os.Stdout).WithAttrs(attrs),
		})
	} else {
		return slog.New(&defaultHandler{
			Handler: handlerOptions.NewTextHandler(os.Stdout).WithAttrs(attrs),
		})
	}
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
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

		return a
	}

	// Converting log level to lowercase
	if a.Key == slog.LevelKey {
		a.Value = slog.StringValue(strings.ToLower(a.Value.String()))
		return a
	}

	// Changing key from "source" to "caller"
	// For now it is commented, if you want to use key name as "caller", just uncomment next 4 lines
	// if a.Key == slog.SourceKey {
	// 	a.Key = "caller"
	// 	return a
	// }

	return a
}
