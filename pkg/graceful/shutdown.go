package graceful

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
)

type (
	ShutdownFn        func(context.Context) error
	TriggerShutdownFn func()
)

// Shutdown calls the given shutdown() function when SIGINT, SIGTERM, SIGHUP or SIGQUIT
// signal is received by the program or when the returned triggerShutdown() function is called.
//
// Returns wait channel that blocks until shutdown() finishes.
// Returns shutdown() function that can be used to trigger the graceful shutdown from within the app.
func Shutdown(
	shutdown ShutdownFn,
	timeout time.Duration,
) (wait chan struct{}, triggerShutdown TriggerShutdownFn) {
	sig := make(chan os.Signal, 1)

	wait = make(chan struct{})
	triggerShutdown = func() {
		sig <- syscall.SIGUSR1
	}

	go func(wait chan struct{}) {
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

		// Blocks until SIGINT, SIGTERM, SIGHUP or SIGQUIT is received.
		<-sig

		// Stop handling subsequent signals. If we receive second SIGTERM (^C), terminate the program.
		signal.Stop(sig)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		slog.Info("graceful: shutdown: shutting down...")

		if err := shutdown(ctx); err != nil {
			slog.ErrorCtx(ctx, fmt.Sprintf("graceful: failed to shutdown(): %v", err))
		}
		slog.Info("graceful: shutdown(): finished")

		close(wait)
	}(wait)

	return wait, triggerShutdown
}

// ShutdownHTTPServer shuts down HTTP server gracefully when SIGINT, SIGTERM, SIGHUP or SIGQUIT
// signal is received by the program or when the returned triggerShutdown() function is called.
//
// When triggered, it closes all listeners and idle connections, and waits for active connections to finish.
//
// Returns wait channel that blocks until all active HTTP connections are finished.
// Returns shutdown() function that can be used to trigger the graceful shutdown from within the app.
func ShutdownHTTPServer(
	srv *http.Server,
	timeout time.Duration,
) (wait chan struct{}, shutdown TriggerShutdownFn) {
	return Shutdown(srv.Shutdown, timeout)
}
