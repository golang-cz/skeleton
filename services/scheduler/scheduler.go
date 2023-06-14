package scheduler

import (
	"fmt"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/graceful"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
)

type Scheduler struct {
	Config *config.AppConfig
}

func New(conf *config.AppConfig, shutdown graceful.TriggerShutdownFn) (*Scheduler, error) {

	//NATS
	if _, err := nats.Connect("api", conf.NATS, shutdown); err != nil {
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

func (app *Scheduler) Close() {
	slog.Info("API: closing NATS & DB connections..")

	nats.Close()
}
