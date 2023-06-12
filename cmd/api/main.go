package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/api"
	apiHttp "github.com/golang-cz/skeleton/services/api/http"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "etc/config.toml", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	// Read config.toml file
	file, err := os.Open(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load and parse config file
	conf, err := config.NewFromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	// Setup application
	err = core.SetupApp(conf, "Skeleton-API", version.VERSION)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              conf.Port,
		Handler:           apiHttp.Router(),
		IdleTimeout:       60 * time.Second, // idle connections
		ReadHeaderTimeout: 10 * time.Second, // request header
		ReadTimeout:       5 * time.Minute,  // request body
		WriteTimeout:      5 * time.Minute,  // response body
		MaxHeaderBytes:    1 << 20,          // 1 MB
	}

	wait, shutdown := graceful.ShutdownHTTPServer(srv, time.Minute)

	// Create app & connect to DB, NATS etc.
	app, err := api.New(conf, shutdown)
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	slog.Info(
		fmt.Sprintf(
			"running application in %s environment version %s",
			api.App.Config.Environment.String(),
			version.VERSION,
		),
	)

	slog.Info(fmt.Sprintf("API serving at %v", api.App.Config.Port))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-wait
}
