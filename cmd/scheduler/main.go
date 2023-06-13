package main

import (
	"flag"
	"fmt"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
	"github.com/golang-cz/skeleton/pkg/version"
	apiHttp "github.com/golang-cz/skeleton/services/api/http"

	"log"
	"net/http"
	"os"
	"time"
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

	//NATS
	if _, err := nats.Connect("scheduler", conf.NATS, shutdown); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		log.Fatal(slogger.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvSchedulerHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		log.Fatal(slogger.ErrorCause(err).Error())
	}

	<-wait
}
