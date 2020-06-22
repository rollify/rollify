package dice

import (
	"context"
	"time"

	"github.com/rollify/rollify/internal/model"
)

// RollerMetricsRecorder knows how to record Roller metrics.
type RollerMetricsRecorder interface {
	// MeasureDiceRollQuantity measures the quantity of dice in a dice roll.
	MeasureDiceRollQuantity(ctx context.Context, rollerType string, diceRoll *model.DiceRoll)
	// MeasureDieRollResult measures the result of a die roll.
	MeasureDieRollResult(ctx context.Context, rollerType string, dieRoll *model.DieRoll)
}

type measuredRoller struct {
	rType string
	rec   RollerMetricsRecorder
	next  Roller
}

// NewMeasureRoller wraps a roller and measures.
func NewMeasureRoller(rollerType string, rec RollerMetricsRecorder, next Roller) Roller {
	return &measuredRoller{
		rType: rollerType,
		rec:   rec,
		next:  next,
	}
}

func (m measuredRoller) Roll(ctx context.Context, dr *model.DiceRoll) error {
	defer func() {
		m.rec.MeasureDiceRollQuantity(ctx, m.rType, dr)
		for _, d := range dr.Dice {
			m.rec.MeasureDieRollResult(ctx, m.rType, &d)
		}

	}()
	return m.next.Roll(ctx, dr)
}

// ServiceMetricsRecorder knows how to record Service metrics.
type ServiceMetricsRecorder interface {
	// ObserveDiceServiceOpDuration increments the metrics  of a dice roll
	ObserveDiceServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration)
}
