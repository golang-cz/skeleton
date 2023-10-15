package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/api/rest"
	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type API struct {
	Config *config.AppConfig
	DB     *data.Database
	HTTP   *http.Server

	shutdownFinished chan struct{}
}

func New(conf *config.AppConfig) (*API, error) {
	// Database
	database, err := data.NewDBSession(conf.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}

	// NATS
	if _, err := nats.Connect("api", conf.NATS); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvAPIHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	app := &API{
		Config: conf,
		DB:     database,

		shutdownFinished: make(chan struct{}, 1),
	}

	restServer := &rest.Server{
		Config: conf,
		DB:     database,
	}

	app.HTTP = &http.Server{
		Addr:              conf.Port,
		Handler:           restServer.Router(),
		IdleTimeout:       60 * time.Second, // idle connections
		ReadHeaderTimeout: 10 * time.Second, // request header
		ReadTimeout:       5 * time.Minute,  // request body
		WriteTimeout:      5 * time.Minute,  // response body
		MaxHeaderBytes:    1 << 20,          // 1 MB
	}

	return app, nil
}

func (app *API) Run() error {
	slog.Info(fmt.Sprintf("API serving at %v", app.Config.Port),
		"env", app.Config.Environment.String(),
		"version", version.VERSION,
	)

	err := app.HTTP.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listening and serving: %w", err)
	}

	// Server is gracefully shutting down. Wait for active
	// connections to be finished before returning.
	<-app.shutdownFinished

	return nil
}

func (app *API) Stop(maxDuration time.Duration) (err error) {
	slog.Info("API: HTTP server gracefully shutting down..", "maxDuration", maxDuration)

	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	start := time.Now()
	defer func() {
		if err != nil {
			slog.Error("API: HTTP server graceful shutdown failed", "duration", time.Since(start), "error", err)
		} else {
			slog.Info("API: HTTP server graceful shutdown finished", "duration", time.Since(start))
		}
	}()

	// Close connections to DB, NATS etc.
	defer app.teardown()

	// Finally, unblock app.Run().
	defer close(app.shutdownFinished)

	// Finish active connections.
	err = app.HTTP.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}

	return nil
}

func (app *API) teardown() {
	slog.Info("API: tearing down..")

	_ = app.DB.Session.Close()
	nats.Close()
}
