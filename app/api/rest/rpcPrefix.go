package rest

import (
	"net/http"
	"strings"
)

// Unlike http.StripPrefix(), this method trims everything before the
// first match of the given path.
//
// stripPrefixBefore("/rpc/")
//
// ==> /api/rpc/Skeleton/GetUser
func stripPrefixBefore(match string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if i := strings.Index(r.URL.Path, match); i > 0 {
				r.URL.Path = r.URL.Path[i:]
			}
			next.ServeHTTP(w, r)
		})
	}
}
