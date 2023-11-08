package rpc

import (
	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type Rpc struct {
	Config *config.AppConfig
	DB     *data.Database
}
