package event

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// Notifier knows how to notify events.
type Notifier interface {
	NotifyDiceRollCreated(ctx context.Context, d model.DiceRoll) error
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Notifier

// Subscriber knows how to subscribe to events.
type Subscriber interface {
	SubscribeDiceRollCreated(roomID, userID string, h func(context.Context, model.DiceRoll) error) error
	UnsubscribeDiceRollCreated(roomID, userID string) error
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Subscriber
