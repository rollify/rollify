package dice

import (
	"context"
	"time"

	"github.com/rollify/rollify/internal/model"
)

// RollerMetricsRecorder knows how to record Roller metrics.
type RollerMetricsRecorder interface {
	MeasureDiceRollQuantity(ctx context.Context, rollerType string, diceRoll *model.DiceRoll)
	MeasureDieRollResult(ctx context.Context, rollerType string, dieRoll *model.DieRoll)
}

//go:generate mockery --case underscore --output dicemock --outpkg dicemock --name RollerMetricsRecorder

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
	MeasureDiceServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output dicemock --outpkg dicemock --name ServiceMetricsRecorder

type measuredService struct {
	rec  ServiceMetricsRecorder
	next Service
}

// NewMeasureService wraps a service and measures.
func NewMeasureService(rec ServiceMetricsRecorder, next Service) Service {
	return &measuredService{
		rec:  rec,
		next: next,
	}
}

func (m measuredService) ListDiceTypes(ctx context.Context) (resp *ListDiceTypesResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceServiceOpDuration(ctx, "ListDiceTypes", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.ListDiceTypes(ctx)
}

func (m measuredService) CreateDiceRoll(ctx context.Context, r CreateDiceRollRequest) (resp *CreateDiceRollResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceServiceOpDuration(ctx, "CreateDiceRoll", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateDiceRoll(ctx, r)
}

func (m measuredService) ListDiceRolls(ctx context.Context, r ListDiceRollsRequest) (resp *ListDiceRollsResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceServiceOpDuration(ctx, "ListDiceRolls", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.ListDiceRolls(ctx, r)
}

func (m measuredService) SubscribeDiceRollCreated(ctx context.Context, r SubscribeDiceRollCreatedRequest) (resp *SubscribeDiceRollCreatedResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceServiceOpDuration(ctx, "SubscribeDiceRollCreated", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.SubscribeDiceRollCreated(ctx, r)
}
