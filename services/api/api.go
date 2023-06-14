package api

import (
	"fmt"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data/database"
)

var App *API

type API struct {
	Config    *config.AppConfig
	DbSession *data.Database
}

func New(conf *config.AppConfig, shutdown graceful.TriggerShutdownFn) (*API, error) {
	// Database
	if _, err := data.NewDBSession(conf.DB); err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}

	//NATS
	if _, err := nats.Connect("api", conf.NATS, shutdown); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvAPIHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	App = &API{Config: conf, DbSession: data.DB}

	return App, nil
}

func (app *API) Close() {
	slog.Info("API: closing NATS & DB connections..")

	App.DbSession.Session.Close()
	nats.Close()
}
