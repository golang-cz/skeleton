package alert

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/version"
	"golang.org/x/exp/slog"
)

func Register(dsn string, environment config.Environment) error {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	// SIT1 + Production environments
	if environment.IsProduction() {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: environment.String(),
			Release:     version.VERSION,
			Transport:   sentrySyncTransport,
		}); err != nil {
			return fmt.Errorf("failed to initialize sentry alerting in production: %w", err)
		}

		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTags(map[string]string{
				"app":         path.Base(os.Args[0]),
				"environment": environment.String(),
			})
		})

		return nil
	}

	// LOCAL / TESTING environment
	sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: environment.String(),
		DebugWriter: os.Stderr,
		Release:     version.VERSION,
		Debug:       true,
	})
	username := ""
	if user, err := user.Current(); err == nil && user.Username != "" {
		username = user.Username
	} else {
		username, _ = os.LookupEnv("USER")
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{
			"app":         path.Base(os.Args[0]),
			"username":    username,
			"environment": environment.String(),
		})
	})

	return nil
}

// Errorf annotates the error, sends event to sentry and logs as error level.
//
// Do not pass nil error, instead use Msgf.
//
// If no context is available pass context.Background().
func Errorf(ctx context.Context, err error, format string, args ...interface{}) error {
	return sendEvent(ctx, fmt.Errorf(format+": %w", append(args, err)...))
}

// Msgf is similar to Errorf, but used when no error is available directly. See Errorf
// for more info.
func Msgf(ctx context.Context, format string, args ...interface{}) error {
	return sendEvent(ctx, toErr(format, args...))
}

func toErr(format string, args ...interface{}) error {
	// Format with Sprint, Sprintf, or neither.
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	return errors.New(msg)
}

// callers can add tags to the event as needed.
func newSentryEvent(err error) *sentry.Event {
	ev := sentry.NewEvent()
	ev.Level = sentry.LevelError
	ev.Message = err.Error()

	ev.Extra = map[string]interface{}{
		"Version":      runtime.Version(),
		"NumCPU":       runtime.NumCPU(),
		"GOMAXPROCS":   runtime.GOMAXPROCS(0),
		"NumGoroutine": runtime.NumGoroutine(),
	}

	stack := sentry.ExtractStacktrace(err)
	ev.Exception = []sentry.Exception{{
		Type:       err.Error(),
		Value:      reflect.TypeOf(err).String(),
		Stacktrace: stack,
	}}

	if c, ok := err.(interface{ Cause() error }); ok {
		e := c.Cause()

		ev.Exception[0].Type = e.Error()
		ev.Exception[0].Value = reflect.TypeOf(e).String()
	}

	return ev
}

func sendEvent(ctx context.Context, err error) error {
	hub := sentry.CurrentHub().Clone()

	// staticcheck fails pipeline if a caller passes nil where context.Context is expected.
	// This sanity check ensures if someone adds a cutom lint ignore (SA1012) then we won't have a
	// panic later on if nil is passed anyways.
	if ctx == nil {
		ctx = context.Background()
	}

	slog.Log(ctx, slog.LevelError, "mes")

	ev := newSentryEvent(err)

	// ev.Tags["firstCtxField"] = ctx.Value("firstCtxField").(string)
	tag := "myOwnTag"
	ev.Tags[tag] = tag
	fmt.Println(ev.Tags)

	hub.CaptureEvent(ev)

	return err
}

// Panic is used within the recover() call in the logger middleware. Do not use this method.
func Panic(rec interface{}) {
	hub := sentry.CurrentHub().Clone()

	err, ok := rec.(error)
	if !ok {
		// if panic() is called with a non-error, we don't get anything except a string in Sentry.
		// No stack, no context, nothing. Just a plain message.
		// hub.Recover is assuming input will always be an error, and only then a stack is generated.
		err = fmt.Errorf("%v", rec)
	}

	if err == http.ErrAbortHandler {
		// panic(http.ErrAbortHandler) is triggered in stdlib for example
		// in http.ReverseProxy handler when the client disconnects etc.
		//
		// We don't want to report these errors to Sentry, since http.Server
		// treats them as special values too and suppresses their stacktrace.
		return
	}

	hub.Recover(err)
}
