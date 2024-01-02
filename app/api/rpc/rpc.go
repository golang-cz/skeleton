package rpc

import (
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
)

type Rpc struct {
	Config *config.Config
	DB     *data.Database
}
