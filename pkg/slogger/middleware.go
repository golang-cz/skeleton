package slogger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/golang-cz/skeleton/internal/sanitize"
	"golang.org/x/exp/slog"
)

type ctxField string

func SloggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		scheme := scheme(r)
		host := host(r)

		// Filter out PIIs from request URL - i dont know how to implement this right now, it use sanitize TODO
		urlQueryString := sanitize.FilterPIIParams(r.URL.Query())
		requestPath := r.URL.Path
		if len(urlQueryString) > 0 {
			requestPath = fmt.Sprintf("%s?%s", requestPath, urlQueryString.Encode())
		}

		refererURL, err := url.Parse(r.Referer())
		if err != nil {
			refererURL = &url.URL{}
		}
		refererURL.RawQuery = sanitize.FilterPIIParams(refererURL.Query()).Encode()

		uri := fmt.Sprintf("%s://%s", scheme, host)

		requestStart := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {

			statusCode := ww.Status()
			timeTaken := time.Since(requestStart)
			requestBodyLength := int(r.ContentLength)
			responseBodyLength := ww.BytesWritten()
			logLevel := statusLevel(statusCode)

			msg := fmt.Sprintf("HTTP %d (%v): %s %s", statusCode, timeTaken, r.Method, uri)

			slog.LogAttrs(ctx, logLevel, msg,
				slog.String("verb", r.Method),
				slog.String("scheme", scheme),
				slog.String("fqdn", host),
				slog.String("request", requestPath),
				slog.String("clientip", r.RemoteAddr),
				slog.String("useragent", r.UserAgent()),
				// Here is used key "referrer", but the word itself should be "referer"? https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referer
				slog.String("referrer", refererURL.String()),
				slog.String("querystring", urlQueryString.Encode()),
				slog.String("uri", uri),
				slog.Int("status", statusCode),
				slog.Int("time_taken", int(timeTaken.Milliseconds())),
				slog.Int("cs_bytes", requestBodyLength),
				slog.Int("sc_bytes", responseBodyLength))
		}()

		next.ServeHTTP(ww, r.WithContext(ctx))
	})
}

// Helper functions
func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	return "http"
}

func host(r *http.Request) string {
	// not standard, but most popular
	host := r.Header.Get("X-Forwarded-Host")
	if host != "" {
		return host
	}

	// RFC 7239
	host = r.Header.Get("Forwarded")
	_, _, host = parseForwarded(host)
	if host != "" {
		return host
	}

	return r.Host
}

func parseForwarded(forwarded string) (addr, proto, host string) {
	if forwarded == "" {
		return
	}

	for _, forwardedPair := range strings.Split(forwarded, ";") {
		if tv := strings.SplitN(forwardedPair, "=", 2); len(tv) == 2 {
			token, value := tv[0], tv[1]
			token = strings.TrimSpace(token)
			value = strings.TrimSpace(strings.Trim(value, `"`))
			switch strings.ToLower(token) {
			case "for":
				addr = value
			case "proto":
				proto = value
			case "host":
				host = value
			}
		}
	}

	return
}

func statusLevel(status int) slog.Level {
	switch {
	case status < 400 && g_disableHandlerSuccessLog:
		return slog.LevelDebug
	case status < 400:
		return slog.LevelInfo
	case status >= 400 && status < 500:
		return slog.LevelWarn
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// extractor of context values
func ctxExtractor(ctx context.Context, ctxField ctxField) (string, bool) {
	value, exists := ctx.Value(ctxField).(string)
	return value, exists
}

// Custom slog handler for extracting values from context
func (h *defaultHandler) Handle(ctx context.Context, r slog.Record) error {
	var ctxField ctxField = "vctraceid"
	slogField := "vctraceid"

	if myField, exists := ctxExtractor(ctx, ctxField); exists {
		r.AddAttrs(slog.String(string(slogField), myField))
	}

	return h.Handler.Handle(ctx, r)
}
