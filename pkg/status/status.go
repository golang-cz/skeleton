package status

import (
	"context"
	"errors"
)

type Probe interface {
	Run(context.Context) Result
}

type Result struct {
	Status        ProbeStatus `json:"status"`
	Info          string      `json:"info"`
	InstanceCount int         `json:"instance_count"`
}

type ProbeStatus uint

const (
	ProbeStatusUnknown ProbeStatus = iota
	ProbeStatusError
	ProbeStatusWarning
	ProbeStatusHealthy
)

var probeStatuses = []string{
	"unknown",
	"error",
	"warning",
	"healthy",
}

// MarshalText implements TextMarshaller.
func (p ProbeStatus) MarshalText() ([]byte, error) {
	for i := range probeStatuses {
		if i == int(p) {
			return []byte(probeStatuses[i]), nil
		}
	}
	return nil, errors.New("could not find status")
}

// String returns the probe status as a string value.
func (p ProbeStatus) String() string {
	buf, _ := p.MarshalText()
	return string(buf)
}

// UnmarshalText implements TextUnmarshaller.
func (p *ProbeStatus) UnmarshalText(text []byte) error {
	*p = ProbeStatusUnknown
	enum := string(text)
	for i, k := range probeStatuses {
		if enum == k {
			*p = ProbeStatus(i)
			return nil
		}
	}
	return nil
}
