package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-cz/looper"
)

func (s *Scheduler) RegisterLooperJobs(ctx context.Context) error {
	interval := time.Duration(s.Config.Looper.Interval)
	waitAfterError := time.Duration(s.Config.Looper.WaitAfterError)
	timeout := time.Duration(s.Config.Looper.JobTimeout)

	jobs := []*looper.Job{
		{
			Name:             "delete-stuck-recordings",
			JobFn:            s.testJob,
			Timeout:          timeout,
			WaitAfterSuccess: interval,
			WaitAfterError:   waitAfterError,
			WithLocker:       true,
		},
	}

	for _, j := range jobs {
		err := s.looper.AddJob(ctx, j)
		if err != nil {
			return fmt.Errorf("add job to looper: %w", err)
		}
	}

	return nil
}

func (s *Scheduler) RegisterGocronJobs() (err error) {
	run := s.gocron
	run.SingletonModeAll()

	return nil
}

func (s *Scheduler) testJob(ctx context.Context) error {
	slog.Debug("test")
	return nil
}
