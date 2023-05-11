package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/alert"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"golang.org/x/exp/slog"
)

func SetupApp(conf *config.AppConfig, appName, version string) error {
	if appName == "" {
		return errors.New("appName is not defined")
	}

	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}

	time.Local = utcLocation

	isProdEnvironment := conf.Environment.IsProduction()

	level := slog.LevelDebug
	if isProdEnvironment {
		level = slog.LevelInfo
	}

	slConf := slogger.Config{
		AppName:                  appName,
		Level:                    level,
		Production:               isProdEnvironment,
		Version:                  version,
		DisableHandlerSuccessLog: conf.DisableHandlerSuccessLog,
	}

	if err := slogger.Register(slConf); err != nil {
		return fmt.Errorf("failed to setup log: %w", err)
	}

	err = alert.Register(conf.Sentry.DSN, conf.Environment)
	if err != nil {
		return fmt.Errorf("failed to init sentry: %w", err)
	}

	return nil
}
