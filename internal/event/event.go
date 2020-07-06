package event

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// Notifier knows how to notify events.
type Notifier interface {
	NotifyDiceRollCreated(ctx context.Context, e model.EventDiceRollCreated) error
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Notifier

// Subscriber knows how to subscribe to events.
type Subscriber interface {
	SubscribeDiceRollCreated(subscribeID, roomID string, h func(context.Context, model.EventDiceRollCreated) error) error
	UnsubscribeDiceRollCreated(subscribeID, roomID string) error
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Subscriber
