package slogger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"golang.org/x/exp/slog"
)

type Config struct {
	AppName                  string
	Version                  string
	Level                    slog.Leveler
	DisableHandlerSuccessLog bool
}

type JSONHandler struct {
	slog.Handler
}

var config Config

var defaultAttrs []slog.Attr

var handlerOptions slog.HandlerOptions

func RegisterProd(slConf Config) error {
	config = slConf
	handlerOptions = slog.HandlerOptions{
		AddSource:   true,
		Level:       slConf.Level,
		ReplaceAttr: replaceAttr,
	}

	defaultAttrs = []slog.Attr{
		slog.String("app", slConf.AppName),
	}

	return nil
}

func RegisterDev(slConf Config) error {
	config = slConf
	handlerOptions = slog.HandlerOptions{
		AddSource:   true,
		Level:       slConf.Level,
		ReplaceAttr: replaceAttr,
	}

	defaultAttrs = []slog.Attr{
		slog.String("app", slConf.AppName),
	}

	return nil
}

func Slogger() *slog.Logger {
	handler := &JSONHandler{
		Handler: handlerOptions.NewJSONHandler(os.Stdout),
	}

	return slog.New(handler)
}

// extractor of context values
func firstValueFromCtx(ctx context.Context) (string, bool) {
	value, exists := ctx.Value(firstCtxField).(string)
	return value, exists
}

// Custom slog handler for extracting values from context
func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	if myField, exists := firstValueFromCtx(ctx); exists {
		r.AddAttrs(slog.String(string(firstCtxField), myField))
	}

	return h.Handler.Handle(ctx, r)
}

type CtxField string

var (
	firstCtxField CtxField = "firstCtxField"
)

// logging middleware
func SloggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, firstCtxField, "first test")
		scheme := scheme(r)
		host := host(r)

		// Filter out PIIs from request URL - i dont know how to implement this right now, it use sanitize TODO
		urlQueryString := r.URL.Query()
		requestPath := r.URL.Path
		if len(urlQueryString) > 0 {
			requestPath = fmt.Sprintf("%s?%s", requestPath, urlQueryString.Encode())
		}

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

			if logLevel == slog.LevelInfo && config.DisableHandlerSuccessLog {
				return
			}

			mes := fmt.Sprintf("HTTP %d (%v): %s %s", statusCode, timeTaken, r.Method, uri)

			// Todo part

			attrs := []slog.Attr{
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
			}

			attrs = append(attrs, defaultAttrs...)

			slogger := slog.New(&JSONHandler{
				Handler: handlerOptions.NewJSONHandler(os.Stdout).WithAttrs(attrs),
			})

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

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
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

	// Changing key from "source" to "caller"
	// For now it is commented, if you want to use key name as "caller", just uncomment next 4 lines
	// if a.Key == slog.SourceKey {
	// 	a.Key = "caller"
	// 	return a
	// }

	return a
}
