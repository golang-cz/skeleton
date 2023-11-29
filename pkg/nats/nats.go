package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"

	"github.com/golang-cz/skeleton/config"
)

var DefaultClient NATSClient = &nopClient{}

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

func Connect(service string, conf config.NATS) (*Client, error) {
	client, err := New(service, conf)
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
	err := DefaultClient.Ping()
	if err != nil {
		return fmt.Errorf("nats ping: %w", err)
	}
	return nil
}

func Stats() nats.Statistics {
	return DefaultClient.Stats()
}

func Close() {
	DefaultClient.Close()
}

func SubscribeCoreNATS(subj string, cb interface{}) error {
	err := DefaultClient.Subscribe(subj, cb)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	return nil
}

func PublishCoreNATS(subj string, v interface{}) error {
	err := DefaultClient.Publish(subj, v)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	return nil
}
