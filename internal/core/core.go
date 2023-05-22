package core

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/alert"
	"github.com/golang-cz/skeleton/pkg/slogger"
)

func SetupApp(conf *config.AppConfig, appName, version string) error {
	if appName == "" {
		return errors.New("appName is not defined")
	}

	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return fmt.Errorf("load utc location: %w", err)
	}
	time.Local = utcLocation

	// Setting default logger
	logger := slogger.Register(appName, conf.Environment.IsProduction())
	slog.SetDefault(logger)
	if err := alert.Register(conf.Sentry.DSN, conf.Environment); err != nil {
		return fmt.Errorf("failed to init sentry: %w", err)
	}

	return nil
}
