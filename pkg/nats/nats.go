package nats

import (
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/nats-io/nats.go"
)

var (
	DefaultClient NATSClient = &nopClient{}
)

type NATSClient interface {
	Conn() *nats.Conn
	Ping() error
	Stats() nats.Statistics
	Unsubscribe()
	Close()

	// Publish a messages to NATS
	Publish(subject string, payload interface{}) error

	// Subscribes to a NATS subject
	Subscribe(subject string, payload interface{}) error
}

func Connect(service string, conf config.NATSConfig, shutdown graceful.TriggerShutdownFn) (*Client, error) {
	client, err := New(service, conf, shutdown)
	if err != nil {
		return nil, err
	}

	DefaultClient = client

	return client, nil
}

func Conn() *nats.Conn {
	return DefaultClient.Conn()
}

func Ping() error {
	return DefaultClient.Ping()
}

func Stats() nats.Statistics {
	return DefaultClient.Stats()
}

func Close() {
	DefaultClient.Close()
}

func SubscribeCoreNATS(subj string, cb interface{}) error {
	return DefaultClient.Subscribe(subj, cb)
}

func PublishCoreNATS(subj string, v interface{}) error {
	return DefaultClient.Publish(subj, v)
}
