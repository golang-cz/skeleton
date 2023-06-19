package main

import (
	"flag"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/scheduler"

	"log"
	"os"
	"time"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"context"
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
	err = core.SetupApp(conf, "Skeleton-Scheduler", version.VERSION)
	if err != nil {
		log.Fatal(err)
	}

	var app *scheduler.Scheduler

	stopListening := func(ctx context.Context) error {
		if app != nil {
			app.Close()
		}
		return nil
	}

	wait, shutdown := graceful.Shutdown(stopListening, time.Minute)

	if app, err = scheduler.New(conf, shutdown); err != nil {
		log.Fatal(err)
	}

	<-wait
}
