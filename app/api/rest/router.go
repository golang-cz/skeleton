package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/alert"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/rs/cors"
)

func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RealIP)
	r.Use(slogger.SloggerMiddleware)
	r.Use(middleware.Recoverer)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: config.App.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r.Use(corsHandler.Handler)

	r.Get("/robots.txt", robots)
	r.Get("/sentry", sentry)
	r.Get("/favicon.ico", favicon)
	r.Mount("/debug/pprof", s.PprofRouter())

	r.Route("/api", func(r chi.Router) {
		r.Get("/status", s.StatusPage)
		r.Mount("/user", s.UserRouter())
		r.Mount("/users", s.UsersRouter())
	})

	return r
}

func robots(w http.ResponseWriter, r *http.Request) {
	// Disallow all robots. We don't want to be indexed by Google etc.
	fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
}

func sentry(w http.ResponseWriter, r *http.Request) {
	if err := alert.Msgf(r.Context(), "request to sentry test endpoint on /sentry", errors.New("panika")); err != nil {
		w.Write([]byte(fmt.Sprintf("Sentry message sent - %s\n", time.Now())))
		return
	}
	w.Write([]byte(fmt.Sprintf("Sentry did not sent - %s\n", time.Now())))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(""))
}
