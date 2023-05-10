package slogger

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"golang.org/x/exp/slog"
)

type JSONHandler struct {
	slog.Handler
}

func Slogger() *slog.Logger {
	opts := slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Converting time of log to UTC format
			if a.Key == slog.TimeKey {
				inputLayout := "2006-01-02 15:04:05.999999999 -0700 MST"
				outputLayout := "2006-01-02T15:04:05Z0700"

				inputTime, err := time.Parse(inputLayout, a.Value.String())
				if err != nil {
					panic(err)
				}

				outputTime := inputTime.UTC()
				a.Value = slog.StringValue(outputTime.Format(outputLayout))

				return a
			}

			// Converting log level to lowercase
			if a.Key == slog.LevelKey {
				a.Value = slog.StringValue(strings.ToLower(a.Value.String()))
				return a
			}

			// changing key from "source" to "caller"
			if a.Key == slog.SourceKey {
				a.Key = "caller"
				return a
			}

			return a
		},
	}

	handler := &JSONHandler{
		Handler: opts.NewJSONHandler(os.Stdout),
	}

	slogger := slog.New(handler)

	return slogger

}

// logging middleware
func SloggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		scheme := scheme(r)
		host := host(r)

		// Filter out PIIs from request URL - i dont know how to implement this right now, it use sanitize TODO
		urlQueryString := r.URL.Query()
		requestPath := r.URL.Path
		// if len(urlQueryString) > 0 {
		requestPath = fmt.Sprintf("%s?%s", requestPath, urlQueryString.Encode())
		// }

		refererURL, err := url.Parse(r.Referer())
		if err != nil {
			refererURL = &url.URL{}
		}
		// refererURL.RawQuery = sanitize.FilterPIIParams(refererURL.Query()).Encode()

		uri := fmt.Sprintf("%s://%s", scheme, host)

		requestStart := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {

			timeTaken := time.Since(requestStart)
			responseBodyLength := ww.BytesWritten()
			requestBodyLength := int(r.ContentLength)
			statusCode := ww.Status()
			logLevel := statusLevel(statusCode)

			mes := fmt.Sprintf("HTTP %d (%v): %s %s", statusCode, timeTaken, r.Method, uri)

			handler := slog.With(
				slog.String("verb", r.Method),
				slog.String("scheme", scheme),
				slog.String("fqdn", host),
				slog.String("request", requestPath),
				slog.String("clientip", r.RemoteAddr),
				slog.String("useragent", r.UserAgent()),
				slog.String("referrer", refererURL.String()),
				slog.String("querystring", urlQueryString.Encode()),
				slog.String("uri", uri),
				slog.Int("status", statusCode),
				slog.Int("time_taken", int(timeTaken.Milliseconds())),
				slog.Int("cs_bytes", requestBodyLength),
				slog.Int("sc_bytes", responseBodyLength),
			).Handler()

			customHandler := &JSONHandler{Handler: handler}
			slogger := slog.New(customHandler)
			slogger.LogAttrs(ctx, logLevel, mes)

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
