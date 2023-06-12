package nats

import (
	"github.com/golang-cz/skeleton/config"
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

	// Publishes a message to a NATS streaming subject
	Publish(subj string, v interface{}) error

	// Subscribes to a NATS streaming subject
	// Defaults to regular subscriptions
	Subscribe(subj string, cb interface{}) error

	// Creates a NATS streaming subscriber queue
	// If multiple clients subscribe to the same subject in a queue group,
	// when a message is published it is only delivered to a single client
	// Defaults to durable subscriptions
	QueueSubscribe(subj string, cb interface{}) error

	// Publish a messages to NATS
	PublishCoreNATS(subj string, cb interface{}) error

	// Subscribes to a NATS subject
	SubscribeCoreNATS(subj string, cb interface{}) error
}

// MessagingModel is the type we use to represent the message semantic of a subscriber.
type MessagingModel int

// MessagingModel types.
const (
	PubSub MessagingModel = iota // 0
	Queue                        // 1
)

var messagingModels = []string{
	"pubsub", "queue",
}

func Connect(service string, conf config.NATSConfig) (*Client, error) {
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
	return DefaultClient.Ping()
}

func Stats() nats.Statistics {
	return DefaultClient.Stats()
}

func Close() {
	DefaultClient.Close()
}

func Publish(subj string, v interface{}) error {
	return DefaultClient.Publish(subj, v)
}

func Subscribe(subj string, cb interface{}) error {
	return DefaultClient.Subscribe(subj, cb)
}

func QueueSubscribe(subj string, cb interface{}) error {
	return DefaultClient.QueueSubscribe(subj, cb)
}

func SubscribeCoreNATS(subj string, cb interface{}) error {
	return DefaultClient.SubscribeCoreNATS(subj, cb)
}

func PublishCoreNATS(subj string, v interface{}) error {
	return DefaultClient.PublishCoreNATS(subj, v)
}
