package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	data "github.com/golang-cz/skeleton/data/database"
)

func (s *Server) UsersRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", s.getUsers)

	return r
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	var users []data.UserStore
	err := s.DB.Session.SQL().SelectFrom("users").All(&users)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
