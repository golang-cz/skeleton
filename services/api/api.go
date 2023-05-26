package api

import (
	"fmt"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type API struct {
	Config    *config.AppConfig
	DbSession *data.Database
}

func New(conf *config.AppConfig) (*API, error) {
	// Database
	database, err := data.NewDBSession(conf.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
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
