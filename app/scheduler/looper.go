package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/pkg/slogger"
)

func looperBeforeJobRuns(conf *config.Config) func(jobName string) {
	level := slog.LevelInfo
	if !conf.Debug.SchedulerJobs {
		level = slogger.LevelTrace
	}

	return func(jobName string) {
		slog.LogAttrs(
			context.Background(),
			level,
			"starting job",
			slog.String("job", jobName),
			slog.String("service", "looper"),
		)
	}
}

func looperWhenJobReturnsNoError(conf *config.Config) func(jobName string, duration time.Duration) {
	level := slog.LevelInfo
	if !conf.Debug.SchedulerJobs {
		level = slogger.LevelTrace
	}

	return func(jobName string, duration time.Duration) {
		slog.LogAttrs(
			context.Background(),
			level,
			"job finished successfully",
			slog.String("job", jobName),
			slog.String("service", "looper"),
			slog.Duration("duration", duration),
		)
	}
}

func looperWhenJobReturnsError(jobName string, duration time.Duration, err error) {
	ctx := context.Background()
	slog.LogAttrs(
		ctx,
		slog.LevelError,
		slogger.ErrorCauseString(err, "job failed"),
		slog.String("job", jobName),
		slog.String("service", "looper"),
		slog.Duration("duration", duration),
		slog.Any("error", err),
	)
}
