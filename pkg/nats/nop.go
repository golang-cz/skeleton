package nats

import (
	"errors"
	"fmt"
	"slog"

	"github.com/nats-io/nats.go"
)

type nopClient struct {
	Alert bool
}

func (c *nopClient) Conn() *nats.Conn { return nil }

func (c *nopClient) Ping() error { return errors.New("nats: nop") }

func (c *nopClient) Stats() nats.Statistics { return nats.Statistics{} }

func (c *nopClient) Unsubscribe() { return }

func (c *nopClient) Close() { return }

func (c *nopClient) Publish(subject string, v interface{}) error {
	err := fmt.Errorf("Trying to publish message to subject (%s) but NATS client is disconnected - payload: %+v", subject, v)
	if c.Alert {
		slog.Error(err.Error())
	} else {
		// Just log a warning to indicate that some functionality depends on NATS but the client is not connected when running in development mode
		slog.Warn("%v", err)
	}
	return nil
}

func (c *nopClient) Subscribe(subject string, payload interface{}) error { return nil }
