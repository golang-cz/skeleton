package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	data "github.com/golang-cz/skeleton/data/database"
	"github.com/golang-cz/skeleton/services/api"
)

type Api struct {
	App *api.API
}

func Router(api *api.API) http.Handler {
	a := &Api{App: api}

	r := chi.NewRouter()

	r.Get("/", a.getUsers)

	return r
}

func (a *Api) getUsers(w http.ResponseWriter, r *http.Request) {
	var users []data.UserStore
	dbsess := a.App.DbSession.Session
	err := dbsess.SQL().SelectFrom("users").All(&users)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
