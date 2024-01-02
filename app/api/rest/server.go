package rest

import (
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
)

type Server struct {
	Config *config.Config
	DB     *data.Database
}
