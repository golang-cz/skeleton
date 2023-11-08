package slogger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/reqctx"
	"github.com/golang-cz/skeleton/internal/sanitize"
)

func SloggerMiddleware(conf *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme := scheme(r)
			host := host(r)

			ctx := r.Context()
			ctx, attrStorage := reqctx.NewAttrStorage(ctx)

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

			uri := fmt.Sprintf("%s://%s%s", scheme, host, requestPath)

			requestStart := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			var reqBody bytes.Buffer
			if conf.Debug.HttpRequestBody {
				r.Body = io.NopCloser(io.TeeReader(r.Body, &reqBody))
			}

			var respBody bytes.Buffer
			if conf.Debug.HttpResponseBody {
				ww.Tee(&respBody)
			}

			defer func() {
				statusCode := ww.Status()
				timeTaken := time.Since(requestStart)
				requestBodyLength := int(r.ContentLength)
				responseBodyLength := ww.BytesWritten()

				logLevel := statusLevel(statusCode)
				if conf.DisableHandlerSuccessLog && statusCode < 400 {
					logLevel = slog.LevelDebug
				}

				msg := fmt.Sprintf("HTTP %d (%v): %s %s", statusCode, timeTaken, r.Method, uri)

				var attrs []slog.Attr
				for key, value := range attrStorage {
					attrs = append(attrs, slog.Any(key, value))
				}

				attrs = append(attrs,
					slog.String("useragent", r.UserAgent()),
					slog.String("referer", refererURL.String()),
				)

				if conf.Environment.IsProduction() {
					attrs = append(attrs,
						slog.String("verb", r.Method),
						slog.String("scheme", scheme),
						slog.String("fqdn", host),
						slog.String("request", requestPath),
						slog.String("clientip", r.RemoteAddr),
						slog.String("querystring", urlQueryString.Encode()),
						slog.String("uri", uri),
						slog.Int("status", statusCode),
						slog.Int("time_taken", int(timeTaken.Milliseconds())),
						slog.Int("cs_bytes", requestBodyLength),
						slog.Int("sc_bytes", responseBodyLength),
					)
				}

				if conf.Debug.HttpRequestBody {
					// Make sure to read full request body if the handler didn't do so.
					io.Copy(io.Discard, r.Body)
					attrs = append(attrs, slog.String("requestBody", reqBody.String()))
				}

				if conf.Debug.HttpResponseBody {
					attrs = append(attrs, slog.String("responseBody", respBody.String()))
				}

				slog.LogAttrs(ctx, logLevel, msg, attrs...)
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

// Helper functions
func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}

	// X-Forwarded-Proto (XFP) header is a de-facto standard header for identifying the
	// protocol (HTTP or HTTPS) that a client used to connect to your proxy or load balancer.
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}

	return "http"
}

func host(r *http.Request) string {
	// X-Forwarded-Host (XFH) header is a de-facto standard header for identifying
	// the original host requested by the client in the Host HTTP request header.
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return host
	}

	// Forwarded header defined by RFC 7239.
	_, _, host := parseForwarded(r.Header.Get("Forwarded"))
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
	case status >= 400 && status < 500:
		return slog.LevelWarn
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Custom slog handler for extracting values from context
func (h *ProductionHandler) Handle(ctx context.Context, r slog.Record) error {
	// Log reqctx values.
	if userId := reqctx.GetUserId(ctx); !userId.IsNil() {
		r.AddAttrs(slog.Any("userId", userId))
	}
	if applicationId := reqctx.GetApplicationId(ctx); !applicationId.IsNil() {
		r.AddAttrs(slog.Any("applicationId", applicationId))
	}

	err := h.Handler.Handle(ctx, r)
	if err != nil {
		return fmt.Errorf("handle record: %w", err)
	}

	return nil
}
