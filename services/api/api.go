package api

import (
	"fmt"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data/database"
)

var App *API

type API struct {
	Config    *config.AppConfig
	DbSession *data.Database
}

func New(conf *config.AppConfig) (*API, error) {
	// Database
	if _, err := data.NewDBSession(conf.DB); err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}
	App = &API{Config: conf, DbSession: data.DB}

	return App, nil
}

func (app *API) Close() {
	slog.Info("API: closing NATS & DB connections..")

	App.DbSession.Session.Close()
}
