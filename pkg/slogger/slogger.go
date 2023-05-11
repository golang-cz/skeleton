package slogger

import (
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type Config struct {
	AppName                  string
	Level                    slog.Leveler
	Production               bool
	Version                  string
	DisableHandlerSuccessLog bool
}

type DefaultHandler struct {
	slog.Handler
}

var (
	g_disableHandlerSuccessLog bool
)

func Register(slConf Config) error {
	g_disableHandlerSuccessLog = slConf.DisableHandlerSuccessLog

	handlerOptions := slog.HandlerOptions{
		AddSource:   true,
		Level:       slConf.Level,
		ReplaceAttr: replaceAttr,
	}

	defaultAttrs := []slog.Attr{
		slog.String("app", slConf.AppName),
	}

	slog.SetDefault(textHandler(handlerOptions, defaultAttrs))
	if slConf.Production {
		slog.SetDefault(jsonHandler(handlerOptions, defaultAttrs))
	}

	return nil
}

func jsonHandler(handlerOptions slog.HandlerOptions, attrs []slog.Attr) *slog.Logger {
	return slog.New(&DefaultHandler{
		Handler: handlerOptions.NewJSONHandler(os.Stdout).WithAttrs(attrs),
	})
}

func textHandler(handlerOptions slog.HandlerOptions, attrs []slog.Attr) *slog.Logger {
	return slog.New(&DefaultHandler{
		Handler: handlerOptions.NewTextHandler(os.Stdout).WithAttrs(attrs),
	})
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
