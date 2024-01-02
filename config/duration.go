package config

import (
	"fmt"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	parsedDuration, err := time.ParseDuration(string(b))
	if err != nil {
		return fmt.Errorf("parse duration %s: %w", string(b), err)
	}

	*d = Duration(parsedDuration)
	return nil
}
