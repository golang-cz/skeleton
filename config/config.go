package config

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

var App AppConfig

// DebugMode show all http requests which our application does in curl format
var DebugMode = false

type AppConfig struct {
	AllowedOrigins           []string    `toml:"allowed_origins"`
	Port                     string      `toml:"bind_address"`
	DB                       DBConfig    `toml:"db"`
	Goose                    GooseConfig `toml:"goose"`
	Sentry                   Sentry      `toml:"sentry"`
	Environment              Environment `toml:"environment"`
	DisableHandlerSuccessLog bool        `toml:"disable_handler_success_log"`
}

type Sentry struct {
	DSN string `toml:"dsn"`
}

// DBConfig represents the convo database configurations that can be found in config.toml or config.sample.toml
type DBConfig struct {
	AppName           string `toml:"app_name"`
	MaxConnectionLife string `toml:"conn_max_lifetime"`
	ConnectionTimeout int    `toml:"connect_timeout"`
	Database          string `toml:"database"`
	DebugQueries      bool   `toml:"debug_queries"`
	Host              string `toml:"host"`
	MaxIdleConns      int    `toml:"max_idle_conns"`
	MaxOpenConns      int    `toml:"max_open_conns"`
	ReadOnly          bool   `toml:"read_only"`
	Username          string `toml:"username"`
	Password          string `toml:"password"`
	SSLMode           string `toml:"sslmode"`
	ReportQueryErrors bool   `toml:"report_query_errors"`

	// IsMigration should be updated at runtime, used by goose migrations
	IsMigration bool `yaml:"-"`
}

type GooseConfig struct {
	Dir    string `toml:"dir"`
	Driver string `toml:"driver"`
}

func NewFromReader(content io.Reader) (*AppConfig, error) {
	_, err := toml.NewDecoder(content).Decode(&App)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config content: %w", err)
	}

	return &App, nil
}
