package status

import (
	"context"
	"encoding/json"
	"fmt"
	"slog"
	"strings"
	"time"

	"github.com/golang-cz/skeleton/pkg/nats"

	natsio "github.com/nats-io/nats.go"
)

type HealthProbe struct {
	Subject string
}

func (p *HealthProbe) Run(_ context.Context) Result {
	// creates NATS inbox where services can reply back with their healthz status
	replyInbox := natsio.NewInbox()

	if err := nats.Ping(); err != nil {
		return Result{
			Status: ProbeStatusError,
			Info:   fmt.Errorf("failed to ping nats: %w", err).Error(),
		}
	}

	// sets up a sync subscriber so all service healthz replies can be read sequentially
	sub, err := nats.Conn().SubscribeSync(replyInbox)
	if err != nil {
		return Result{
			Status: ProbeStatusError,
			Info:   fmt.Errorf("failed to subscribe to inbox: %w", err).Error(),
		}
	}

	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			err = fmt.Errorf("failed to  unsubscribe from inbox: %w", err)
			slog.Error(err.Error())
		}
	}()

	// publishes a message to the service healthz subscriber with a temporary inbox address waiting for the replies
	if err := nats.PublishCoreNATS(p.Subject, ServiceStats{ReplyInbox: replyInbox}); err != nil {
		return Result{
			Status: ProbeStatusError,
			Info:   fmt.Errorf("failed to send ping request: %w", err).Error(),
		}
	}

	replies := []*ServiceStats{}
	for {
		// if we don't get back a reply within 1 second we stop listening
		msg, _ := sub.NextMsg(1 * time.Second)
		if msg == nil {
			break
		}
		var reply *ServiceStats
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			return Result{
				Status: ProbeStatusError,
				Info:   fmt.Errorf("failed to unmarshal reply: %w", err).Error(),
			}
		}
		replies = append(replies, reply)
	}

	info := make([]string, 0, len(replies))
	for _, r := range replies {
		info = append(info, r.String())
	}

	status := ProbeStatusHealthy
	if len(replies) < 1 {
		status = ProbeStatusError
	}

	return Result{
		Status:        status,
		Info:          strings.Join(info, "<br>"),
		InstanceCount: len(replies),
	}
}

var _ Probe = &HealthProbe{}
