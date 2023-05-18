package httpUsers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/golang-cz/skeleton/data/database"
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
		fmt.Printf("User from DB\n")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
