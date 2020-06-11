package storage

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// DiceRollRepository is the repository interface that implementations need to
// implement to manage dice rolls in storage.
type DiceRollRepository interface {
	// CreateDiceRoll creates a new dice roll.
	// If the dice roll doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the dice roll already exists it returns a internalerrors.AlreadyExists error kind.
	CreateDiceRoll(ctx context.Context, r model.DiceRoll) error
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name DiceRollRepository

// RoomRepository is the repository interface that implementations need to
// implement to manage rooms in storage.
type RoomRepository interface {
	// CreateRoom creates a new room.
	// If the room doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the room already exists it returns a internalerrors.AlreadyExists error kind.
	CreateRoom(ctx context.Context, r model.Room) error
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name RoomRepository