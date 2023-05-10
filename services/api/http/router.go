package apiHttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/slogger"

	"github.com/rs/cors"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RealIP)
	r.Use(slogger.SloggerMiddleware)

	if config.App.Environment.IsLocal() {
		r.Use(middleware.Recoverer)
	} else {
		r.Use(middleware.Recoverer)
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: config.App.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-FE-Route"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r.Use(corsHandler.Handler)

	r.Get("/robots.txt", robots)

	r.Get("/", status)

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(""))
	})

	return r
}

func robots(w http.ResponseWriter, r *http.Request) {
	// Disallow all robots. We don't want to be indexed by Google etc.
	fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Server running - %s\n", time.Now())))
}
