package migration

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(args []string, conf *config.Config) error {
	db, err := data.NewDBSession(conf.DB)
	if err != nil {
		return fmt.Errorf("connect to DB: %w", err)
	}

	fmt.Println(conf.DB)
	goose.SetBaseFS(migrations)

	err = goose.SetDialect(conf.Goose.Driver)
	if err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	collectedMigrations, err := goose.CollectMigrations("migrations", int64(0), int64((1<<63)-1))
	if err != nil {
		return fmt.Errorf("collect migrations: %w", err)
	}

	// We don't allow "timestamped" migration version numbers to be run outside of "local" / "test" environment.
	// Timestamp migration get fixed by CI pipeline and gets renamed to sequential order
	if conf.Environment.IsProduction() {
		for _, m := range collectedMigrations {
			// If the version is bigger than 20000000000000, we assume it's a "timestamp" version, which shouldn't be merged in.
			if m.Version >= 20000000000000 {
				return fmt.Errorf("cannot run timestamped migration %q on non-dev environment", m)
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
		if len(args[1:]) == 1 {
			return fmt.Errorf("missing name for migration")
		}

		if len(args[1:]) > 2 {
			return fmt.Errorf("too many arguments %v", args[1:])
		}
	}

	for {
		err := goose.Run(cmd, db.Driver().(*sql.DB), dir, args[1:]...)
		if errors.Is(err, goose.ErrNoNextVersion) || errors.Is(err, goose.ErrNoCurrentVersion) {
			return nil
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

			return fmt.Errorf("running migration: %w", err)
		}

		if !loop {
			break
		}

		// New DB session for each loop. Fixes upper/db cache bug after schema changes.
		db.Close()

		db, err = data.NewDBSession(conf.DB)
		if err != nil {
			return fmt.Errorf("connect to DB: %w", err)
		}

	}

	return nil
}
