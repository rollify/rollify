package prometheus

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	gohttpmetrics "github.com/slok/go-http-metrics/metrics"
	gohttpmetricsprom "github.com/slok/go-http-metrics/metrics/prometheus"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/model"
)

const prefix = "rollify"

// Types used to avoid collisions with the same interface naming.
type httpRecorder gohttpmetrics.Recorder

// Recorder satisfiies multiple interfaces related with metrics measuring
// it will implement Prometheus based metrics backend.
type Recorder struct {
	httpRecorder

	diceRollQuantity *prometheus.HistogramVec
	dieRollResult    *prometheus.CounterVec
}

// NewRecorder returns a new recorder implementation for prometheus.
func NewRecorder(reg prometheus.Registerer) Recorder {
	r := Recorder{
		httpRecorder: gohttpmetricsprom.NewRecorder(gohttpmetricsprom.Config{Registry: reg}),

		diceRollQuantity: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "dice_roller",
			Name:      "dice_roll_die_quantity",
			Help:      "The quantity of dice on dice rolls.",
			Buckets:   []float64{1, 2, 5, 8, 12, 18, 25, 40, 60, 100},
		}, []string{"roller_type"}),

		dieRollResult: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Subsystem: "dice_roller",
			Name:      "die_roll_results_total",
			Help:      "The total number of die rolls.",
		}, []string{"roller_type", "die_type", "die_side"}),
	}

	reg.MustRegister(
		r.diceRollQuantity,
		r.dieRollResult,
	)

	return r
}

// MeasureDiceRollQuantity satisfies dice.RollerMetricsRecorder.
func (r Recorder) MeasureDiceRollQuantity(ctx context.Context, rollerType string, diceRoll *model.DiceRoll) {
	r.diceRollQuantity.WithLabelValues(rollerType).Observe(float64(len(diceRoll.Dice)))
}

// MeasureDieRollResult satisfies dice.RollerMetricsRecorder.
func (r Recorder) MeasureDieRollResult(ctx context.Context, rollerType string, dieRoll *model.DieRoll) {
	r.dieRollResult.WithLabelValues(rollerType, dieRoll.Type.ID(), strconv.Itoa(int(dieRoll.Side))).Inc()
}

var (
	_ apiv1.MetricsRecorder      = Recorder{}
	_ dice.RollerMetricsRecorder = Recorder{}
)
