package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-cz/skeleton/app/api/rest"
	"github.com/golang-cz/skeleton/app/api/rpc"
	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/proto"
)

type API struct {
	Config *config.AppConfig
	DB     *data.Database
	HTTP   *http.Server
	RPC    *rpc.Rpc

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

	rpcSever := &rpc.Rpc{
		Config: conf,
		DB:     database,
	}

	rpcHandler := proto.NewSkeletonServer(rpcSever)

	app := &API{
		Config:           conf,
		DB:               database,
		RPC:              rpcSever,
		shutdownFinished: make(chan struct{}, 1),
	}

	restServer := &rest.Server{
		Config: conf,
		DB:     database,
	}

	app.HTTP = &http.Server{
		Addr:              conf.Port,
		Handler:           restServer.Router(rpcHandler),
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
		slog.Any("env", app.Config.Environment.String()),
		slog.Any("version", version.VERSION),
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
	slog.Info("API: HTTP server gracefully shutting down..", slog.Any("maxDuration", maxDuration))

	// Unblock app.Run() at the very end (defer calls are executed in LIFO order).
	defer close(app.shutdownFinished)

	// Teardown the app (close DB, NATS connections etc).
	defer app.teardown()

	// Log graceful shutdown duration/error.
	start := time.Now()
	defer func() {
		if err != nil {
			slog.Error("API: HTTP server graceful shutdown failed", slog.Any("duration", time.Since(start)), slog.Any("error", err))
		} else {
			slog.Info("API: HTTP server graceful shutdown finished", slog.Any("duration", time.Since(start)))
		}
	}()

	// Within given maxDuration.
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	// Finish all active connections.
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
