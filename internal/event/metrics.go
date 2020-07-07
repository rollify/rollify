package event

import (
	"context"
	"time"

	"github.com/rollify/rollify/internal/model"
)

// NotifierMetricsRecorder knows how to measure Notifier.
type NotifierMetricsRecorder interface {
	MeasureNotifyOpDuration(ctx context.Context, notifierType, op string, success bool, t time.Duration)
}

type measuredNotifier struct {
	notifierType string
	rec          NotifierMetricsRecorder
	next         Notifier
}

// NewMeasuredNotifier wraps a Notifier and measures.
func NewMeasuredNotifier(notifierType string, rec NotifierMetricsRecorder, next Notifier) Notifier {
	return &measuredNotifier{
		notifierType: notifierType,
		rec:          rec,
		next:         next,
	}
}

func (m measuredNotifier) NotifyDiceRollCreated(ctx context.Context, e model.EventDiceRollCreated) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureNotifyOpDuration(ctx, m.notifierType, "NotifyDiceRollCreated", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.NotifyDiceRollCreated(ctx, e)
}
