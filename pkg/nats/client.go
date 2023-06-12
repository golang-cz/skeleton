package nats

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/lg"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/url"
	"reflect"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
	"github.com/nats-io/stan.go"
)

type Client struct {
	// Connection to the core NATS server
	NATSConn *nats.Conn

	// Connection to the NATS Streaming server
	STANConn stan.Conn

	// Name of service on the network.
	Service string

	stanSubs []stan.Subscription
}

func New(service string, conf config.NATSConfig, shutdown graceful.TriggerShutdownFn) (*Client, error) {
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

	if conf.Cluster == "" {
		return nil, errors.New("missing required cluster argument")
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

	stanOpts := []stan.Option{
		stan.NatsConn(client.NATSConn),
		stan.ConnectWait(5 * time.Second),
		stan.PubAckWait(2 * time.Minute),
		stan.SetConnectionLostHandler(func(conn stan.Conn, err error) {
			// Sleep for 0-60s randomly, so we don't shutdown all pods at once.
			timeout := time.Duration(rand.Intn(6)) * 10 * time.Second
			log.Warn().Msgf("%s: NATS-Streaming connection lost: triggering shutdown() in %v", service, timeout)
			time.Sleep(timeout)

			shutdown()
		}),
		stan.Pings(5, 2), // Ping every 5 seconds. Connection will be considered lost after 2 failed Ping attempts.
	}

	// Connect to NATS-Streaming
	for i := 1; i <= 10; i++ {
		client.STANConn, err = stan.Connect(conf.Cluster, fmt.Sprintf("%s-%s", service, nuid.Next()), stanOpts...)
		if err == nil {
			break
		}
		log.Warn().Msgf("failed to connect to NATS-Streaming: retry [%v/%v]", i, 10)
	}

	if err != nil {
		err = fmt.Errorf("failed to connect to NATS-Streaming: %w", err)
		log.Fatal().Err(err).Msg(lg.ErrorCause(err).Error())
	}

	return client, nil
}

func (c *Client) Conn() stan.Conn {
	return c.STANConn
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
	for _, sub := range c.stanSubs {
		sub.Unsubscribe()
	}

	// NOTE: Should we unsubscribe from NATS subscriptions too?
	// Let's not for now -- keep the HEALTH subscription up.
}

// Close grecefuly shutdown NATS and NATS-Streaming connections
func (c *Client) Close() {
	log.Info().Msgf("nats: closing STAN connection")
	c.STANConn.Close()

	log.Info().Msgf("nats: closing NATS connection")
	c.NATSConn.Close()
}

func (c *Client) Publish(subj string, v interface{}) error {
	// Log alert if message is trying to be published when NATS client is disconnected
	if !c.NATSConn.IsConnected() {
		err := fmt.Errorf("%s: trying to publish message to subject %q but NATS client is disconnected", c.Service, subj)
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	}
	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			err = fmt.Errorf("%s: failed to acknowledge message %q of subject %q: %w", c.Service, ackedNuid, subj, err)
			log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
		}
	}
	var err error
	var nuid string
	switch data := v.(type) {
	case []byte:
		nuid, err = c.STANConn.PublishAsync(subj, data, ackHandler)
		if err != nil {
			return fmt.Errorf("%s: error publishing message to %s - nuid %s: %w", c.Service, subj, nuid, err)
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		nuid, err = c.STANConn.PublishAsync(subj, b, ackHandler)
		if err != nil {
			return fmt.Errorf("%s: error publishing message to %s - nuid %s: %w", c.Service, subj, nuid, err)
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
	sub, err := c.STANConn.Subscribe(subj, func(msg *stan.Msg) {
		if err := msg.Ack(); err != nil {
			return
		}
		go processMsg(msg.Subject, msg.Data, argType, cb)
	}, stan.SetManualAckMode())
	if err != nil {
		return fmt.Errorf("failed to subscribe to %q: %w", subj, err)
	}
	c.stanSubs = append(c.stanSubs, sub)
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
	sub, err := c.STANConn.QueueSubscribe(subj, c.Service, func(msg *stan.Msg) {
		if err := msg.Ack(); err != nil {
			return
		}
		go processMsg(msg.Subject, msg.Data, argType, cb)
	}, stan.SetManualAckMode(), stan.DurableName(fmt.Sprintf("%s-%s", c.Service, subj)))
	if err != nil {
		return fmt.Errorf("failed to subscribe to %q: %w", subj, err)
	}
	c.stanSubs = append(c.stanSubs, sub)
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
