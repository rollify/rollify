package event

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// Notifier knows how to notify events.
type Notifier interface {
	// SendDiceRollCreated will notify that a user created a new dice roll.
	SendDiceRollCreated(ctx context.Context, d model.DiceRoll) error
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Notifier

// Subscriber knows how to subscribe to events.
type Subscriber interface {
	// RecvDiceRollCreated subscribes to DiceRollCreated using a handler.
	RecvDiceRollCreated(roomID string, h func(model.DiceRoll) error)
}

//go:generate mockery -case underscore -output eventmock -outpkg eventmock -name Subscriber
