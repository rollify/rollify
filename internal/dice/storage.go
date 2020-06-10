package dice

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// Repository is the repository interface that implementations need to
// implement to manage dice in storage.
type Repository interface {
	// CreateDiceRoll creates a new dice roll.
	// If the dice roll doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the dice roll already exists it returns a internalerrors.AlreadyExists error kind.
	CreateDiceRoll(ctx context.Context, r model.DiceRoll) error
}

//go:generate mockery -case underscore -output dicemock -outpkg dicemock -name Repository
