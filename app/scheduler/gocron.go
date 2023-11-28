package scheduler

import (
	"context"
	"log/slog"

	"github.com/go-co-op/gocron"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/slogger"
)

func (s *Scheduler) gocronRegisterJobEventListeners(conf *config.Config) error {
	for _, job := range s.gocron.Jobs() {
		job.RegisterEventListeners(
			gocron.BeforeJobRuns(gocronBeforeJobRuns(conf)),
			gocron.WhenJobReturnsNoError(gocronWhenJobReturnsNoError(conf)),
			gocron.WhenJobReturnsError(gocronWhenJobReturnsError),
		)
	}

	return nil
}

func gocronBeforeJobRuns(conf *config.Config) func(jobName string) {
	level := slog.LevelInfo
	if conf.Environment.IsLocal() && !conf.Debug.SchedulerJobs {
		level = slogger.LevelTrace
	}

	return func(jobName string) {
		slog.LogAttrs(
			context.Background(),
			level,
			"starting job",
			slog.String("job", jobName),
			slog.String("service", "gocron"),
		)
	}
}

func gocronWhenJobReturnsNoError(conf *config.Config) func(jobName string) {
	level := slog.LevelInfo
	if conf.Environment.IsLocal() && !conf.Debug.SchedulerJobs {
		level = slogger.LevelTrace
	}

	return func(jobName string) {
		slog.LogAttrs(
			context.Background(),
			level,
			"job finished successfully",
			slog.String("job", jobName),
			slog.String("service", "gocron"),
		)
	}
}

func gocronWhenJobReturnsError(jobName string, err error) {
	slog.LogAttrs(
		context.Background(),
		slog.LevelError,
		slogger.ErrorCauseString(err, "job failed"),
		slog.String("job", jobName),
		slog.String("service", "gocron"),
		slog.Any("error", err),
	)
}
