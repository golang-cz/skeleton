package status

import (
	"fmt"
	"github.com/golang-cz/skeleton/pkg/version"
	"os"
	"runtime"
)

const megaByte = 1 << (10 * 2)

func GetServiceStats() *ServiceStats {
	stats := &ServiceStats{
		AppVersion:    version.VERSION,
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
		GoVersion:     runtime.Version(),
	}

	runtime.ReadMemStats(&stats.MemStats)
	stats.Hostname, _ = os.Hostname()

	return stats
}

type ServiceStats struct {
	AppVersion string `json:"app_version"`

	NumGoroutines int              `json:"go_routines"`
	GoVersion     string           `json:"go_version"`
	Hostname      string           `json:"hostname"`
	NumCPU        int              `json:"cpu_cores"`
	MemStats      runtime.MemStats `json:"mem_stats"`

	ReplyInbox string `json:"reply_inbox"`
}

func (s *ServiceStats) String() string {
	return fmt.Sprintf("%v (%v), %v, mem: %vM, heap: %v/%vM, goroutines: %v",
		s.AppVersion,
		s.GoVersion,
		s.Hostname,
		s.MemStats.Sys/megaByte,
		s.MemStats.HeapAlloc/megaByte,
		s.MemStats.HeapSys/megaByte,
		s.NumGoroutines,
	)
}
