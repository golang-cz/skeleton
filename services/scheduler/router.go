package scheduler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-cz/skeleton/pkg/slogger"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RealIP)
	r.Use(slogger.SloggerMiddleware)
	r.Use(middleware.Recoverer)

	return r
}
