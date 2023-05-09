package api

import (
	"fmt"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
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

	App = &API{Config: conf}

	return App, nil
}

func (app *API) Close() {
	data.Close()
}
