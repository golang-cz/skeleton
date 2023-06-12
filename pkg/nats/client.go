package nats

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/lg"
	"github.com/rs/zerolog/log"
	"net/url"
	"reflect"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	// Connection to the core NATS server
	NATSConn *nats.Conn

	// Name of service on the network.
	Service string

	natsSubs []*nats.Subscription
}

func New(service string, conf config.NATSConfig) (*Client, error) {
	opts := []nats.Option{
		nats.Timeout(5 * time.Second),
		nats.ReconnectWait(1 * time.Second),
		nats.MaxReconnects(-1), // try reconnecting forever
		nats.DisconnectHandler(func(nc *nats.Conn) {
			// Only alert if connection wasn't previously closed
			if !nc.IsClosed() {
				// Note: nc.LastError() might be nil, don't wrap it
				log.Error().Err(fmt.Errorf("%s: NATS client disconnected", service))
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info().Msgf("%s: NATS client reconnected to %+v", service, nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info().Msgf("%s: NATS client connection closed", service)
		}),
	}

	_, err := url.Parse(conf.Server)
	if err != nil {
		return nil, err
	}

	if service == "" {
		return nil, errors.New("missing required service argument")
	}

	client := &Client{Service: service}

	// Connect to core NATS server so we can configure callbacks
	for i := 1; i <= 10; i++ {
		client.NATSConn, err = nats.Connect(conf.Server, opts...)
		if err == nil {
			break
		}
		log.Warn().Msgf("failed to connect to NATS: retry [%v/%v]", i, 10)
		time.Sleep(time.Second)
	}
	if err != nil {
		err = fmt.Errorf("failed to connect to NATS: %w", err)
		log.Fatal().Err(err).Msg(lg.ErrorCause(err).Error())
	}

	return client, nil
}

func (c *Client) Conn() *nats.Conn {
	return c.NATSConn
}

func (c *Client) Ping() error {
	if c.NATSConn == nil || !c.NATSConn.IsConnected() {
		return errors.New("NATS is disconnected")
	}
	return nil
}

func (c *Client) Stats() nats.Statistics {
	if c.NATSConn == nil {
		return c.NATSConn.Stats()
	}
	return nats.Statistics{}
}

func (c *Client) Unsubscribe() {
	for _, sub := range c.natsSubs {
		sub.Unsubscribe()
	}

	// NOTE: Should we unsubscribe from NATS subscriptions too?
	// Let's not for now -- keep the HEALTH subscription up.
}

// Close grecefuly shutdown NATS and NATS-Streaming connections
func (c *Client) Close() {
	log.Info().Msgf("nats: closing NATS connection")
	c.NATSConn.Close()
}

func (c *Client) Publish(subj string, v interface{}) error {
	// Log alert if message is trying to be published when NATS client is disconnected
	if !c.NATSConn.IsConnected() {
		err := fmt.Errorf("%s: trying to publish message to subject %q but NATS client is disconnected", c.Service, subj)
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	}
	var err error
	switch data := v.(type) {
	case []byte:
		err = c.NATSConn.Publish(subj, data)
		if err != nil {
			return fmt.Errorf("%s: error publishing message to %s: %w", c.Service, subj, err)
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = c.NATSConn.Publish(subj, b)
		if err != nil {
			return fmt.Errorf("%s: error publishing message to %s: %w", c.Service, subj, err)
		}
	}

	return nil
}

// Broadcast. Fan-out. All subscribed clients will get the messages of a given subject.
func (c *Client) Subscribe(subj string, cb interface{}) error {
	// check if callback is valid, expects to be a function with two arguments
	// eg; func PostPublished(subject string, post *presenter.Post)
	argType, _, err := argInfo(cb)
	if err != nil || argType == nil {
		return fmt.Errorf("%s: invalid argument type for callback: %w", c.Service, err)
	}
	// wrap NATS subscribe and provide JSON encoding support for backwards compatibility
	sub, err := c.NATSConn.Subscribe(subj, func(msg *nats.Msg) {
		if err := msg.Ack(); err != nil {
			return
		}
		go processMsg(msg.Subject, msg.Data, argType, cb)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to %q: %w", subj, err)
	}
	c.natsSubs = append(c.natsSubs, sub)
	return nil
}

// Queue. Only a single client (of all subscribed clients) will get the message of a given subject.
func (c *Client) QueueSubscribe(subj string, cb interface{}) error {
	// check if callback is valid, expects to be a function with two arguments
	// eg; func PostPublished(subject string, post *presenter.Post)
	argType, _, err := argInfo(cb)
	if err != nil || argType == nil {
		return fmt.Errorf("%s: invalid argument type for callback: %w", c.Service, err)
	}
	// wrap NATS queue subscribe and provide JSON encoding support for backwards compatibility
	// durable name is a combination if the service + subject, eg; api-data.post.published
	sub, err := c.NATSConn.QueueSubscribe(subj, c.Service, func(msg *nats.Msg) {
		if err := msg.Ack(); err != nil {
			return
		}
		go processMsg(msg.Subject, msg.Data, argType, cb)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to %q: %w", subj, err)
	}
	c.natsSubs = append(c.natsSubs, sub)
	return nil
}

func (c *Client) PublishCoreNATS(subj string, v interface{}) error {
	// Log alert if message is trying to be published when NATS client is disconnected
	if !c.NATSConn.IsConnected() {
		err := fmt.Errorf("Trying to publish message to subject (%s) but NATS client is disconnected - payload: %+v", subj, v)
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	}

	var err error
	switch data := v.(type) {
	case []byte:
		err = c.NATSConn.Publish(subj, data)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = c.NATSConn.Publish(subj, b)
	}

	if err != nil {
		return fmt.Errorf("error publishing message to %s : %w", subj, err)
	}

	return nil
}

func (c *Client) SubscribeCoreNATS(subj string, cb interface{}) error {
	// check if callback is valid, expects to be a function with two arguments
	// eg; func PostPublished(subject string, post *presenter.Post)
	argType, _, err := argInfo(cb)
	if err != nil || argType == nil {
		return fmt.Errorf("invalid argument type for callback: %w", err)
	}

	_, err = c.NATSConn.Subscribe(subj, func(msg *nats.Msg) {
		processMsg(msg.Subject, msg.Data, argType, cb)
	})
	return err
}

// Process NATS published message and unmarshal data into callback argument
func processMsg(msgSubject string, msgData []byte, argType reflect.Type, cb interface{}) {
	var oPtr reflect.Value
	if argType.Kind() != reflect.Ptr {
		oPtr = reflect.New(argType)
	} else {
		oPtr = reflect.New(argType.Elem())
	}

	if msgData != nil && len(msgData) > 0 {
		err := json.Unmarshal(msgData, oPtr.Interface())
		if err != nil {
			return
		}
	}
	if argType.Kind() != reflect.Ptr {
		oPtr = reflect.Indirect(oPtr)
	}
	reflect.ValueOf(cb).Call([]reflect.Value{reflect.ValueOf(msgSubject), oPtr})
}

// Reads callback function and return total number of arguments and their types
func argInfo(cb interface{}) (reflect.Type, int, error) {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		return nil, 0, fmt.Errorf("callback handler needs to be a function")
	}
	numArgs := cbType.NumIn()
	if numArgs != 2 {
		return nil, numArgs, fmt.Errorf("callback handler needs to have 2 arguments")
	}
	return cbType.In(numArgs - 1), numArgs, nil
}
