package status

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/upper/db/v4"

	data "github.com/golang-cz/skeleton/data/database"
	"github.com/golang-cz/skeleton/pkg/status"
	"github.com/golang-cz/skeleton/pkg/ws"
)

type probe struct {
	status.Probe
	Key string `json:"key"`
}

type result struct {
	status.Result
	Key string `json:"key"`
}

var (
	serviceProbes = []probe{
		{
			Key: "Api",
			Probe: &status.HealthProbe{
				Subject: "api",
			},
		},
	}

	uptimeProbes = []probe{
		{
			Key: "SkeletonDb",
			Probe: &status.Postgres{
				GetDB: func() db.Session { return data.DB.Session },
			},
		},
	}
)

func StatusPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	results := run(ctx, append(uptimeProbes, serviceProbes...))

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		ws.JSON(w, 200, results)
		return
	}

	i := len(uptimeProbes) // Helper index to split the slice.
	statusPage, err := status.RenderTemplate(struct {
		Uptime      []result
		ServiceInfo []result
	}{
		Uptime:      results[:i],
		ServiceInfo: results[i:],
	})
	if err != nil {
		ws.RespondError(w, r, 500, fmt.Errorf("failed to render template: %w", err))
		return
	}

	ws.HTML(w, 200, statusPage)
}

func run(ctx context.Context, probes []probe) []result {
	results := make([]result, len(probes))

	var wg sync.WaitGroup
	for i, p := range probes {
		i, p := i, p // Copy for local goroutine.

		wg.Add(1)
		go func() {
			defer wg.Done()

			results[i] = result{
				Key:    p.Key,
				Result: p.Probe.Run(ctx),
			}
		}()
	}
	wg.Wait()

	return results
}
