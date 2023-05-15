package status

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-cz/skeleton/pkg/bond"
)

type Postgres struct {
	GetDB func() bond.Backend
}

var _ Probe = &Postgres{}

func (p *Postgres) Run(_ context.Context) Result {
	db := p.GetDB()

	err := db.Ping()
	if err != nil {
		return Result{
			Status: ProbeStatusError,
			Info:   err.Error(),
		}
	}

	status := ProbeStatusError
	var (
		version        string
		numConnections int64
		maxConnections int64
		connectedApps  string
		uptime         int64
	)
	row, err := db.QueryRow(`
		SELECT
			split_part(version(), ' ', 2),
			(SELECT SUM(numbackends) FROM pg_stat_database),
			(SELECT setting	FROM pg_settings WHERE name='max_connections'),
			array_to_string(array(SELECT application_name || ': ' || COUNT(state='active' OR NULL) || '/' || count(1) FROM pg_stat_activity WHERE application_name <> '' GROUP BY application_name ORDER BY lower(application_name)), ', '),
			extract(epoch FROM current_timestamp - pg_postmaster_start_time())::bigint
		`)
	if err == nil {
		status = ProbeStatusHealthy
		_ = row.Scan(&version, &numConnections, &maxConnections, &connectedApps, &uptime)
	}

	return Result{
		Status: status,
		Info: fmt.Sprintf("PostgreSQL v%v, conns: %v/%v (%v), uptime: %v",
			version,
			numConnections,
			maxConnections,
			connectedApps,
			time.Duration(uptime)*time.Second,
		),
	}
}
