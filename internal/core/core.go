package core

import (
	"fmt"
	"time"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/alert"
)

func SetupApp(conf *config.AppConfig, appName, version string) error {
	utcLocation, err := time.LoadLocation("UTC")
	if err != nil {
		return err
	}

	time.Local = utcLocation

	err = alert.Register(conf.Sentry.DSN, conf.Environment)
	if err != nil {
		return fmt.Errorf("failed to init sentry: %w", err)
	}

	return nil
}
