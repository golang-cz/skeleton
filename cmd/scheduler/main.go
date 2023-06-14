package main

import (
	"flag"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/services/scheduler"

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
		Handler:           scheduler.Router(),
		IdleTimeout:       60 * time.Second, // idle connections
		ReadHeaderTimeout: 10 * time.Second, // request header
		ReadTimeout:       5 * time.Minute,  // request body
		WriteTimeout:      5 * time.Minute,  // response body
		MaxHeaderBytes:    1 << 20,          // 1 MB
	}

	_, shutdown := graceful.ShutdownHTTPServer(srv, time.Minute)

	if _, err := scheduler.New(conf, shutdown); err != nil {
		log.Fatal(err)
	}

	select {}
}
