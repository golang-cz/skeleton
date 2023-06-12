package nats

import (
	"errors"
	"fmt"
	"github.com/golang-cz/skeleton/pkg/lg"
	"github.com/rs/zerolog/log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type nopClient struct {
	Alert bool
}

func (c *nopClient) Conn() stan.Conn { return nil }

func (c *nopClient) Ping() error { return errors.New("nats: nop") }

func (c *nopClient) Stats() nats.Statistics { return nats.Statistics{} }

func (c *nopClient) Unsubscribe() { return }

func (c *nopClient) Close() { return }

// Publish In case the NATS clients can't establish and the fallback noop implementation is used we send alert for every message that is published when running in production mode
func (c *nopClient) Publish(subj string, v interface{}) error {
	err := fmt.Errorf("Trying to publish message to subject (%s) but NATS client is disconnected - payload: %+v", subj, v)
	if c.Alert {
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	} else {
		// Just log a warning to indicate that some functionality depends on NATS but the client is not connected when running in development mode
		log.Warn().Msgf("%v", err)
	}
	return nil
}

func (c *nopClient) Subscribe(subj string, cb interface{}) error { return nil }

func (c *nopClient) QueueSubscribe(subj string, cb interface{}) error { return nil }

func (c *nopClient) PublishCoreNATS(subj string, v interface{}) error {
	err := fmt.Errorf("Trying to publish message to subject (%s) but NATS client is disconnected - payload: %+v", subj, v)
	if c.Alert {
		log.Error().Err(err).Msg(lg.ErrorCause(err).Error())
	} else {
		// Just log a warning to indicate that some functionality depends on NATS but the client is not connected when running in development mode
		log.Warn().Msgf("%v", err)
	}
	return nil
}

func (c *nopClient) SubscribeCoreNATS(subj string, cb interface{}) error { return nil }
