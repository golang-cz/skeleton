package main

import (
	"flag"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/scheduler"

	"log"
	"os"
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

	// Create app & connect to DB, NATS etc.
	app, err := scheduler.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
