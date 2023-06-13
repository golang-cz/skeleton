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

	// NATS streaming has all the core NATS features, plus;
	// -Log based persistence
	// -At-Least-Once Delivery model, giving reliable message delivery
	// -Rate matched on a per subscription basis
	// -Replay/Restart
	// -Last Value Semantics

	// Publish a messages to NATS
	PublishCoreNATS(subj string, cb interface{}) error

	// Subscribes to a NATS subject
	SubscribeCoreNATS(subj string, cb interface{}) error
}

// MessagingModel is the type we use to represent the message semantic of a subscriber.
type MessagingModel int

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
	return DefaultClient.SubscribeCoreNATS(subj, cb)
}

func PublishCoreNATS(subj string, v interface{}) error {
	return DefaultClient.PublishCoreNATS(subj, v)
}
