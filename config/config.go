package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	AllowedOrigins           []string    `toml:"allowed_origins"`
	DisableHandlerSuccessLog bool        `toml:"disable_handler_success_log"`
	Environment              Environment `toml:"environment"`
	Port                     string      `toml:"bind_address"`
	BaseUrl                  string      `toml:"base_url"`

	// Subgroups
	AWS        AWS        `toml:"aws"`
	DB         DB         `toml:"db"`
	Debug      Debug      `toml:"debug"`
	StatusPage StatusPage `toml:"status_page"`
	Looper     Looper     `toml:"looper"`
	Goose      Goose      `toml:"goose"`
	NATS       NATS       `toml:"nats"`
	Redis      Redis      `toml:"redis"`
	Sentry     Sentry     `toml:"sentry"`
}

type Debug struct {
	HttpOutgoingRequests bool `toml:"http_outgoing_requests"`
	HttpRequestBody      bool `toml:"http_request_body"`
	HttpResponseBody     bool `toml:"http_response_body"`
	DBQueries            bool `toml:"db_queries"`
	SchedulerJobs        bool `toml:"scheduler_jobs"`
}

type AWS struct {
	Region     string     `toml:"region"`
	S3         S3         `toml:"s3"`
	CloudWatch CloudWatch `toml:"cloud_watch"`
}

type S3 struct {
	Bucket     string `toml:"bucket"`
	Cloudfront string `toml:"cloudfront"`
	KMSKey     string `toml:"kms_key"`
}

type MediaConvert struct {
	Role     string `toml:"role"`
	Queue    string `toml:"queue"`
	Endpoint string `toml:"endpoint"`
}

type CloudWatch struct {
	LogGroup LogGroup `toml:"log_group"`
}

type LogGroup struct {
	IVS          string `toml:"ivs"`
	MediaConvert string `toml:"media_convert"`
	Transcribe   string `toml:"transcribe"`
}

// DB represents video database configurations that can be found in config.toml or config.sample.toml
type DB struct {
	AppName           string `toml:"app_name"`
	MaxConnectionLife string `toml:"conn_max_lifetime"`
	ConnectionTimeout int    `toml:"connect_timeout"`
	Database          string `toml:"database"`
	Host              string `toml:"host"`
	MaxIdleConns      int    `toml:"max_idle_conns"`
	MaxOpenConns      int    `toml:"max_open_conns"`
	ReadOnly          bool   `toml:"read_only"`
	Username          string `toml:"username"`
	Password          string `toml:"password"`
	SSLMode           string `toml:"sslmode"`
	ReportQueryErrors bool   `toml:"report_query_errors"`
}

type StatusPage struct {
	OrgId  string `toml:"org_id"`
	UserID string `toml:"user_id"`
}

type Looper struct {
	Interval       Duration `toml:"interval"`
	WaitAfterError Duration `toml:"wait_after_error"`
	JobTimeout     Duration `toml:"job_timeout"`
}

type Goose struct {
	Dir    string `toml:"dir"`
	Driver string `toml:"driver"`
}

type NATS struct {
	Server  string `toml:"server"`
	Cluster string `toml:"cluster"`
}

type Redis struct {
	Host string `toml:"host"`
}

type Sentry struct {
	DSN string `toml:"dsn"`
}

func NewFromReader(confFile string) (*Config, error) {
	file, err := os.Open(confFile)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var conf Config
	_, err = toml.NewDecoder(file).Decode(&conf)
	if err != nil {
		return nil, fmt.Errorf("parse config content: %w", err)
	}

	err = validate(&conf)
	if err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &conf, nil
}

func validate(conf *Config) (err error) {
	return nil
}
