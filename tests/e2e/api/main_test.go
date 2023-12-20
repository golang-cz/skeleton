package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/goware/urlx"

	"github.com/golang-cz/skeleton/app/api"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
	"github.com/golang-cz/skeleton/internal/core"
	"github.com/golang-cz/skeleton/pkg/version"
	"github.com/golang-cz/skeleton/proto/client/skeleton"
)

type E2EServices struct {
	ProjectRootDirectory string
	User                 *data.User
	API                  *api.API
	DB                   *data.Database
	Config               *config.Config
	Client               *http.Client
	RPCClient            skeleton.Skeleton
	UserId               uuid.UUID
}

var E2E *E2EServices

func TestMain(m *testing.M) {
	E2E = &E2EServices{}

	// Get the current Go module path.
	projectRootDirectory, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/golang-cz/skeleton").Output()
	if err != nil {
		log.Fatalf("locating current Go module path: %v", err)
	}
	E2E.ProjectRootDirectory = strings.TrimSpace(string(projectRootDirectory))

	// Load config file.
	conf, err := config.NewFromReader(filepath.Join(E2E.ProjectRootDirectory, "etc/test.toml"))
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	E2E.Config = conf // Adding configuration to E2E

	// Setup application
	err = core.SetupApp(conf, "SKELETON-E2E", version.VERSION)
	if err != nil {
		log.Fatalf("setting up app: %v", err)
	}

	err = initDB(conf.DB.Database)
	if err != nil {
		log.Fatalf("starting DB: %v", err)
	}

	// Create app & connect to DB, NATS etc.
	app, err := api.New(context.Background(), conf)
	if err != nil {
		log.Fatalf("creating API: %v", err)
	}
	defer app.Stop(time.Second)

	E2E.DB = app.DB

	internalUrl, _ := urlx.Parse(fmt.Sprintf("http://localhost%s/_api", conf.Port))

	E2E.Client = &http.Client{Timeout: 10 * time.Second}

	E2E.RPCClient = skeleton.NewSkeletonClient(internalUrl.String(), E2E.Client)

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("failed to run API: %v", err)
		}
	}()

	// check if http server is ready via health endpoint
	for {
		// release allocated resources on every call
		// if I did not wrap it in function it would be called at the end of all for iterations which could create
		// memory leaks
		serverReady := func() bool {
			resp, _ := http.Get(fmt.Sprintf("%s/ping", internalUrl))
			if resp != nil {
				defer resp.Body.Close()
			}

			if resp != nil && resp.StatusCode == 200 {
				slog.Info("server is ready on port", "port", conf.Port)
				return true
			}

			slog.Info("waiting for server to be ready....")

			time.Sleep(time.Millisecond * 20)
			return false
		}()

		if serverReady {
			break
		}
	}

	// Run tests.
	os.Exit(m.Run())
}

func initDB(database string) error {
	slog.Debug("Initializing DB", "database", database)

	// Execute the command to create the new schema
	createCmd := exec.Command("bash", "-c", fmt.Sprintf(`docker exec skeleton-postgres /bin/sh -c '/home/db.sh create %s'`, database))
	_, err := createCmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			slog.Error("Failed to create schema", "schema", database, "err", err)
			return fmt.Errorf("failed to create schema %q: %w", database, errors.New(string(e.Stderr)))
		}
		slog.Error("Failed to run db.sh script", "err", err)
		return fmt.Errorf("failed to run db.sh script: %w", err)
	}

	// Import the schema.sql script into the new schema
	importCmd := exec.Command("bash", "-c", fmt.Sprintf(`docker exec skeleton-postgres /bin/sh -c '/home/db.sh import %s %s'`, database, "/home/schema.sql"))
	_, err = importCmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			slog.Error("Failed to import schema.sql for", "schema", database, "err", err)
			return fmt.Errorf("failed to import schema.sql for %q: %w", database, errors.New(string(e.Stderr)))
		}
		slog.Error("Failed to run db.sh script", "err", err)
		return fmt.Errorf("failed to run db.sh script: %w", err)
	}

	slog.Info("Schema created and schema.sql imported", "schema", database)

	return nil
}
