package scheduler

import (
	"fmt"
	"log/slog"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
)

type Scheduler struct {
	Config *config.AppConfig
}

func New(conf *config.AppConfig) (*Scheduler, error) {
	// NATS
	if _, err := nats.Connect("api", conf.NATS); err != nil {
		err = fmt.Errorf("failed to connect to NATS server: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	if err := status.HealthSubscriber(events.EvSchedulerHealth); err != nil {
		err = fmt.Errorf("failed enable health subscribe: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	scheduler := &Scheduler{
		Config: conf,
	}

	return scheduler, nil
}

func (s *Scheduler) Run() error {
	select {} // TODO

	return nil
}

func (s *Scheduler) Stop() {
	slog.Info("API: closing NATS & DB connections..")

	nats.Close()
}
