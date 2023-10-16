package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"slog"

	"github.com/golang-cz/skeleton/pkg/alert"
)

func RespondError(w http.ResponseWriter, r *http.Request, status int, err error) {
	ctx := r.Context()

	resp := map[string]interface{}{
		"error": errorCause(err).Error(),
		// "vctraceid": vctraceid.FromContext(r.Context()),
		"vctraceid": "somethin",
	}

	slog.LogAttrs(ctx, slog.LevelError, err.Error())

	JSON(w, status, resp)
	return
}

// errorCause returns the very first (non-nil) error cause,
// same as errors.Cause(err) from github.com/pkg/errors.
func errorCause(err error) error {
	var cause error
	for e := err; e != nil; e = errors.Unwrap(e) {
		cause = e
	}
	return cause
}

func HTML(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.String:
		w.Write([]byte(v.String()))
	case reflect.Slice:
		w.Write(v.Interface().([]byte))
	default:
		alert.Msgf(context.Background(), "invalid response data type:", v.Kind())
	}
}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	var err error
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(b) > 0 {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(b)
}
