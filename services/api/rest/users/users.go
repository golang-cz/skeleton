package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	data "github.com/golang-cz/skeleton/data/database"
)

func Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", getUsers)

	return r
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []data.UserStore
	dbsess := data.DB.Session
	err := dbsess.SQL().SelectFrom("users").All(&users)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
