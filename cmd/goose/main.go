package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data/migration"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
)

var (
	flags    = flag.NewFlagSet("goose", flag.ExitOnError)
	confFile = flags.String("config", "etc/config.toml", "path to config file")
)

func main() {
	flags.Usage = usage
	flags.Parse(os.Args[1:])

	file, err := os.Open(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse config file.
	conf, err := config.NewFromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	if conf == nil {
		log.Fatal(errors.New("failed to unmarshal config"))
	}

	conf.DB.IsMigration = true

	args := flags.Args()
	if len(args) < 1 {
		log.Fatal("no command provided")
	}

	if args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	err = core.SetupApp(conf, "Skeleton-Migration", version.VERSION)
	if err != nil {
		log.Fatal(err)
	}

	err = migration.RunMigrations(args, conf)
	if err != nil {
		log.Fatal(fmt.Errorf("goose migration: %w", err))
	}
}

func usage() {
	fmt.Print(usagePrefix)
	flags.PrintDefaults()
	fmt.Print(usageCommands)
}

var (
	usagePrefix = `
Usage: goose -config=FILE COMMAND
Options:
`

	usageCommands = `
Commands:
	create MIGRATION_NAME [go|sql] Create new migration	
	up         Migrate the DB to the most recent version available
	down       Roll back the version by 1
	redo       Re-run the latest migration
	status     Dump the migration status for the current DB
	dbversion  Print the current version of the database
`
)
