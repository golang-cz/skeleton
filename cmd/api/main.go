package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/api"
	"golang.org/x/exp/slog"
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

	// Create app & connect to DB, NATS etc.
	app, err := api.New(conf)
	if err != nil {
		log.Fatal(err)
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
