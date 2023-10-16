package rest

import (
	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type Server struct {
	Config *config.AppConfig
	DB     *data.Database
}
