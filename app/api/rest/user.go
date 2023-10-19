package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	data "github.com/golang-cz/skeleton/data/database"
	"github.com/upper/db/v4"
)

func (s *Server) UserRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/{uuid}", func(r chi.Router) {
		r.Use(s.UserCtx)
		r.Get("/detail", getUser)
	})

	return r
}

func (s *Server) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "uuid")
		user, err := s.dbGetUser(userID)
		if err != nil {
			http.Error(w, http.StatusText(418), 418)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) dbGetUser(userID string) (userstore data.UserStore, err error) {
	var user data.UserStore

	err = s.DB.Session.Get(&user, db.Cond{"id": userID})
	if err != nil {
		return user, fmt.Errorf("user from db: %w", err)
	}
	return user, nil
}

func getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(data.UserStore)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
