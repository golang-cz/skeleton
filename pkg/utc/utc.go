package utc

import (
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func NowPtr() *time.Time {
	now := Now()
	return &now
}
