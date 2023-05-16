package main

import (
	"database/sql"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
)

var (
	flags    = flag.NewFlagSet("goose", flag.ExitOnError)
	confFile = flags.String("config", "etc/config.toml", "path to config file")
	//go:embed migrations/*.sql
	migrations embed.FS
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

	db, err := data.NewDBSession(conf.DB)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to main DB: %w", err))
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(conf.Goose.Driver); err != nil {
		log.Fatal(err)
	}

	collectedMigrations, err := goose.CollectMigrations("migrations", int64(0), int64((1<<63)-1))
	if err != nil {
		log.Fatal(err)
	}

	// We don't allow "timestamped" migration version numbers to be run outside of "local" / "test" environment.
	// Timestamp migration get fixed by CI pipeline and gets renamed to sequential order
	if conf.Environment.IsProduction() {
		for _, m := range collectedMigrations {
			// If the version is bigger than 20000000000000, we assume it's a "timestamp" version, which shouldn't be merged in.
			if m.Version >= 20000000000000 {
				log.Fatal(
					fmt.Errorf("cannot run timestamped migration %q on non-dev environment", m),
				)
			}
		}
	}

	cmd := args[0]
	var loop bool
	if cmd == "up" {
		cmd = "up-by-one"
		loop = true
	}

	dir := "migrations"
	if cmd == "create" {
		dir = conf.Goose.Dir
	}

	for {
		err := goose.Run(cmd, db.Driver().(*sql.DB), dir, args[1:]...)
		if err == goose.ErrNoNextVersion || err == goose.ErrNoCurrentVersion {
			break
		}
		if err != nil {
			var e *pq.Error
			if errors.As(err, &e) {
				err = fmt.Errorf(
					"%s: %s(%s) - position(%s) internal-position(%s) internal-query(%s) where(%s) schema(%s) table(%s) column(%s) data-type-name(%s) constraint(%s): %w",
					e.Code,
					e.Message,
					e.Detail,
					e.Position,
					e.InternalPosition,
					e.InternalQuery,
					e.Where,
					e.Schema,
					e.Table,
					e.Column,
					e.DataTypeName,
					e.Constraint,
					err,
				)
			}
			slog.Error(err.Error())
		}
		if !loop {
			break
		}

		// New DB session for each loop. Fixes upper/db cache bug after schema changes.
		// dbsess.Close()
		db.Close()
		sess, err := data.NewDBSession(conf.DB)
		if err != nil {
			slog.Error(err.Error())
		}

		db = sess
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
