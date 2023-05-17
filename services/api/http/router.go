package apiHttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/alert"
	"github.com/golang-cz/skeleton/pkg/slogger"
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
			"Accept", "Authorization", "Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r.Use(corsHandler.Handler)

	r.Get("/robots.txt", robots)
	r.Get("/status", status)
	// r.Get("/_api/status", httpStatus.StatusPage)

	r.Get("/sentry", sentry)

	r.Get("/favicon.ico", favicon)
	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Route("/{uuid}", func(r chi.Router) {
				r.Use(UserCtx)
				r.Get("/detail", getUser)
			})
		})
	})
	return r
}

func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "uuid")
		user, err := dbGetUser(userID)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func dbGetUser(userID string) (user string, err error) {
	return userID, nil
}

func getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(string)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	w.Write([]byte(fmt.Sprintf("user-uuid:%s", user)))
}

func robots(w http.ResponseWriter, r *http.Request) {
	// Disallow all robots. We don't want to be indexed by Google etc.
	fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Server running!!! - %s\n", time.Now())))
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

func allUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(""))
}
