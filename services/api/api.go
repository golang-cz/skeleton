package api

import (
	"fmt"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/lg"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/status"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type API struct {
	Config    *config.AppConfig
	DbSession *data.Database
}

func New(conf *config.AppConfig, shutdown graceful.TriggerShutdownFn) (*API, error) {
	// Database
	database, err := data.NewDBSession(conf.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}

	//NATS + Streaming
	if _, err := nats.Connect("api", conf.NATS, shutdown); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		log.Fatal().Err(err).Msg(lg.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvAPIHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	}

	app := &API{Config: conf, DbSession: database}

	//NATS
	if _, err := nats.Connect("api", conf.NATS, shutdown); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvAPIHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	return app, nil
}

func (app *API) Close() {
	slog.Info("API: closing NATS & DB connections..")

	app.DbSession.Session.Close()
	nats.Close()
}
