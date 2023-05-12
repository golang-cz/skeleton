package slogger

import (
	"os"
	"strings"
	"time"

	"github.com/golang-cz/skeleton/pkg/version"
	"golang.org/x/exp/slog"
)

type Config struct {
	AppName                  string
	Production               bool
	Version                  string
	DisableHandlerSuccessLog bool
}

type defaultHandler struct {
	slog.Handler
}

var g_disableHandlerSuccessLog bool

func Register(slConf Config) error {
	g_disableHandlerSuccessLog = slConf.DisableHandlerSuccessLog

	level := slog.LevelDebug
	if slConf.Production {
		level = slog.LevelInfo
	}

	handlerOptions := slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replaceAttr,
	}

	defaultAttrs := []slog.Attr{
		slog.String("app", slConf.AppName),
		slog.String("release", version.VERSION),
	}

	slog.SetDefault(setDefaultHandler(handlerOptions, defaultAttrs, slConf.Production))

	return nil
}

func setDefaultHandler(handlerOptions slog.HandlerOptions, attrs []slog.Attr, production bool) *slog.Logger {
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
