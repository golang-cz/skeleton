package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/golang-cz/looper"

	"github.com/golang-cz/skeleton/config"
	"github.com/golang-cz/skeleton/data"
	"github.com/golang-cz/skeleton/pkg/events"
	"github.com/golang-cz/skeleton/pkg/nats"
	"github.com/golang-cz/skeleton/pkg/pretty"
	"github.com/golang-cz/skeleton/pkg/slogger"
	"github.com/golang-cz/skeleton/pkg/status"
)

type Scheduler struct {
	Config     *config.Config
	NatsClient *nats.Client
	DB         *data.Database

	gocron  *gocron.Scheduler
	looper  *looper.Looper
	stopped chan struct{}
}

func New(ctx context.Context, conf *config.Config) (*Scheduler, error) {
	// NATS
	natsClient, err := nats.Connect("scheduler", conf.NATS)
	if err != nil {
		err = fmt.Errorf("connect to NATS server: %w", err)
		slog.Error(slogger.ErrorCause(err).Error())
	}

	// Database
	database, err := data.NewDBSession(conf.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}

	err = status.HealthSubscriber(events.EvSchedulerHealth)
	if err != nil {
		err = fmt.Errorf("enable health subscriber: %w", err)
		slog.Error(err.Error())
	}

	// Looper
	looperConfig := looper.Config{
		StartupTime: time.Second * 5,
	}
	loop := looper.New(looperConfig)

	panicHandler := getPanicHandler(conf)
	gocron.SetPanicHandler(panicHandler)
	looper.SetPanicHandler(panicHandler)

	// Gocron
	cron := gocron.NewScheduler(time.UTC)

	schedulerApp := &Scheduler{
		Config:     conf,
		NatsClient: natsClient,
		DB:         database,

		gocron: cron,
		looper: loop,

		stopped: make(chan struct{}, 1),
	}

	return schedulerApp, nil
}

func (s *Scheduler) RegisterJobs(ctx context.Context, conf *config.Config) (err error) {
	s.looper.RegisterHooks(looperBeforeJobRuns(conf), looperWhenJobReturnsNoError(conf), looperWhenJobReturnsError)
	err = s.RegisterLooperJobs(ctx)
	if err != nil {
		return fmt.Errorf("registering looper jobs: %w", err)
	}

	err = s.RegisterGocronJobs()
	if err != nil {
		return fmt.Errorf("registering gocron jobs: %w", err)
	}

	err = s.gocronRegisterJobEventListeners(conf)
	if err != nil {
		return fmt.Errorf("register gocron event listeners: %w", err)
	}

	return nil
}

func (s *Scheduler) Run() {
	s.looper.Start()
	s.gocron.StartAsync()
	s.printRegisteredJobsCount()

	<-s.stopped
}

func (s *Scheduler) RunJob(jobNames ...string) (err error) {
	slog.Info("job names", slog.Any("jobNames", jobNames))
	// gocron must be started before we can run specific jobs with given tag
	s.gocron.StartAsync()

	for _, jobName := range jobNames {
		err = s.gocron.RunByTagWithDelay(jobName, 5*time.Second)
		if err == nil {
			// found in gocron skip to next job
			continue
		}

		if !errors.Is(err, gocron.ErrJobNotFoundWithTag) {
			return fmt.Errorf("start job wit tag %s: %w", jobName, err)
		}

		err = s.looper.StartJobByName(jobName)
		if err != nil {
			return fmt.Errorf("start job by name: %w", err)
		}
	}

	s.printRegisteredJobsCount()
	return nil
}

func (s *Scheduler) printRegisteredJobsCount() {
	slog.Info("scheduler started jobs",
		slog.Any("gocron jobs", len(s.gocron.Jobs())),
		slog.Any("looper jobs", len(s.looper.Jobs())),
	)
}

func (s *Scheduler) LogStats(interval time.Duration) {
	type JobInfo struct {
		Runner      string
		Name        string
		Started     *bool
		Running     bool
		LastRun     time.Time
		Runs        uint64
		SuccessRuns uint64
		ErrorRuns   uint64
	}

	for {
		var jobs []JobInfo
		for _, j := range s.looper.Jobs() {
			job := JobInfo{
				Runner:      "looper",
				Name:        j.Name,
				Running:     j.Running,
				Started:     &j.Started,
				LastRun:     j.LastRun.UTC(),
				Runs:        j.RunCountSuccess + j.RunCountError,
				SuccessRuns: j.RunCountSuccess,
				ErrorRuns:   j.RunCountError,
			}
			jobs = append(jobs, job)
		}

		for _, j := range s.gocron.Jobs() {
			job := JobInfo{
				Runner:      "gocron",
				Name:        j.GetName(),
				Running:     j.IsRunning(),
				LastRun:     j.LastRun(),
				Runs:        uint64(j.RunCount()),
				SuccessRuns: 0,
				ErrorRuns:   0,
			}
			jobs = append(jobs, job)
		}

		attrs := make([]slog.Attr, 0, len(jobs))
		for _, job := range jobs {
			g := slog.Group(fmt.Sprintf("%s - %s", job.Runner, job.Name),
				slog.Any("runner", job.Runner),
				slog.Any("name", job.Name),
				slog.Any("running", job.Running),
				slog.Any("started", job.Started),
				slog.Any("lastRun", job.LastRun),
				slog.Any("runs", job.Runs),
				slog.Any("successRuns", job.SuccessRuns),
				slog.Any("errorRuns", job.ErrorRuns),
			)
			attrs = append(attrs, g)
		}

		slog.LogAttrs(context.Background(), slog.LevelInfo, "scheduler statistics", attrs...)

		time.Sleep(interval)
	}
}

type stopFn struct {
	name string
	fn   func()
}

// blocking operation
// it waits until all jobs are finished
func (s *Scheduler) Stop() {
	slog.Info("stopping scheduler")
	defer close(s.stopped)

	stopFns := []stopFn{
		{"nats", s.NatsClient.Close},
		{"gocron", s.gocron.Stop},
		{"looper", s.looper.Stop},
	}

	var wg sync.WaitGroup
	wg.Add(len(stopFns))
	for _, fn := range stopFns {
		go s.stop(&wg, fn)
	}
	wg.Wait()

	err := s.DB.Close()
	if err != nil {
		slog.Error("close db connection", slog.Any("err", err))
	}

	slog.Info("all scheduler jobs stopped")
}

func (s *Scheduler) stop(wg *sync.WaitGroup, sfn stopFn) {
	defer wg.Done()
	slog.Warn(fmt.Sprintf("%v: stopping", sfn.name))
	sfn.fn()
	slog.Info(fmt.Sprintf("%v: stopped", sfn.name))
}

func getPanicHandler(conf *config.Config) func(jobName string, recoverData interface{}) {
	return func(jobName string, recoverData interface{}) {
		if conf.Environment.IsLocal() {
			slog.Error(fmt.Sprintf("job panicked %v", jobName))
			pretty.Recoverer(recoverData)
		} else {
			slog.Error(fmt.Sprintf("job panicked %v", jobName),
				slog.Any("stack", string(debug.Stack())),
			)
		}
	}
}
