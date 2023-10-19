package rest

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

func (s *Server) PprofRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", pprof.Index)
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Get("/heap", func(w http.ResponseWriter, r *http.Request) {
			pprof.Handler("heap").ServeHTTP(w, r)
		})
	})

	return r
}
