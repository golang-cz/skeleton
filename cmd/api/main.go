package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-cz/skeleton/app/api"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "etc/config.toml", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	// Load and parse config file
	conf, err := config.NewFromReader(*confFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Setup application
	err = core.SetupApp(conf, "skeleton-api", version.VERSION)
	if err != nil {
		log.Fatalf("setup app: %v", err)
	}

	// Create app & connect to DB, NATS etc.
	app, err := api.New(context.Background(), conf)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigs
		slog.Info("received signal", "signal", sig)
		app.Stop(10 * time.Second)
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
