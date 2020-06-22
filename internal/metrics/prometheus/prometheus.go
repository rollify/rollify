package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	gohttpmetrics "github.com/slok/go-http-metrics/metrics"
	gohttpmetricsprom "github.com/slok/go-http-metrics/metrics/prometheus"

	"github.com/rollify/rollify/internal/http/apiv1"
)

// Types used to avoid collisions with the same interface naming.
type httpRecorder gohttpmetrics.Recorder

// Recorder satisfiies multiple interfaces related with metrics measuring
// it will implement Prometheus based metrics backend.
type Recorder struct {
	httpRecorder
}

// NewRecorder returns a new recorder implementation for prometheus.
func NewRecorder(reg prometheus.Registerer) Recorder {
	return Recorder{
		httpRecorder: gohttpmetricsprom.NewRecorder(gohttpmetricsprom.Config{Registry: reg}),
	}
}

var (
	_ apiv1.MetricsRecorder = Recorder{}
)
