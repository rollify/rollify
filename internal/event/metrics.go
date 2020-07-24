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

// SubscriberMetricsRecorder knows how to measure Subscriber.
type SubscriberMetricsRecorder interface {
	MeasureSubscriberSubscribeOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration)
	MeasureSubscriberUnsubscribeOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration)
	MeasureSubscriberEventHandleOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration)
	AddSubscriberQuantity(ctx context.Context, subscriberType, subscription string, quantity int)
}

type measuredSubscriber struct {
	subscriberType string
	rec            SubscriberMetricsRecorder
	next           Subscriber
}

// NewMeasuredSubscriber wraps a Subscriber and measures.
func NewMeasuredSubscriber(subscriberType string, rec SubscriberMetricsRecorder, next Subscriber) Subscriber {
	return &measuredSubscriber{
		subscriberType: subscriberType,
		rec:            rec,
		next:           next,
	}
}

func (m measuredSubscriber) SubscribeDiceRollCreated(ctx context.Context, subscribeID, roomID string, h func(context.Context, model.EventDiceRollCreated) error) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureSubscriberSubscribeOpDuration(ctx, m.subscriberType, "DiceRollCreated", err == nil, time.Since(t0))
	}(time.Now())

	defer func() {
		if err == nil {
			m.rec.AddSubscriberQuantity(ctx, m.subscriberType, "DiceRollCreated", 1)
		}
	}()

	// Wrap also the handler so it measures handle of events.
	measuredHandler := func(ctx context.Context, e model.EventDiceRollCreated) (err error) {
		defer func(t0 time.Time) {
			m.rec.MeasureSubscriberEventHandleOpDuration(ctx, m.subscriberType, "DiceRollCreated", err == nil, time.Since(t0))
		}(time.Now())

		return h(ctx, e)
	}

	return m.next.SubscribeDiceRollCreated(ctx, subscribeID, roomID, measuredHandler)
}

func (m measuredSubscriber) UnsubscribeDiceRollCreated(ctx context.Context, subscribeID, roomID string) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureSubscriberUnsubscribeOpDuration(ctx, m.subscriberType, "DiceRollCreated", err == nil, time.Since(t0))
	}(time.Now())

	defer func() {
		if err == nil {
			m.rec.AddSubscriberQuantity(ctx, m.subscriberType, "DiceRollCreated", -1)
		}
	}()

	return m.next.UnsubscribeDiceRollCreated(ctx, subscribeID, roomID)
}
