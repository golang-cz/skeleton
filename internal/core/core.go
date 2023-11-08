package core

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/slogger"
)

func SetupApp(conf *config.Config, appName, version string) error {
	if appName == "" {
		return errors.New("appName is not defined")
	}

	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return fmt.Errorf("load utc location: %w", err)
	}
	time.Local = utcLocation

	// Set default logger
	logger, err := slogger.New(appName, version, !conf.Environment.IsLocal())
	if err != nil {
		return fmt.Errorf("creating new slog logger: %w", err)
	}
	slog.SetDefault(logger)

	return nil
}
