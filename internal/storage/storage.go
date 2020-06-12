package storage

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// DiceRollList is a list of dice rolls.
type DiceRollList struct {
	Items []model.DiceRoll
}

// ListDiceRollsOpts are the options used by the storage to list dice rolls.
type ListDiceRollsOpts struct {
	RoomID string
	UserID string
}

// DiceRollRepository is the repository interface that implementations need to
// implement to manage dice rolls in storage.
type DiceRollRepository interface {
	// CreateDiceRoll creates a new dice roll.
	// If the dice roomID is empty it returns a internalerrors.NotValid error kind.
	// If the dice userID is empty it returns a internalerrors.NotValid error kind.
	// If the dice roll doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the dice roll already exists it returns a internalerrors.AlreadyExists error kind.
	CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error
	// ListDiceRolls lists dice rolls.
	// If the dice roomID option is empty it returns a internalerrors.NotValid error kind.
	ListDiceRolls(ctx context.Context, opts ListDiceRollsOpts) (*DiceRollList, error)
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name DiceRollRepository

// RoomRepository is the repository interface that implementations need to
// implement to manage rooms in storage.
type RoomRepository interface {
	// CreateRoom creates a new room.
	// If the room doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the room already exists it returns a internalerrors.AlreadyExists error kind.
	CreateRoom(ctx context.Context, r model.Room) error
	// RoomExists returns true if the room exists.
	RoomExists(ctx context.Context, id string) (exists bool, err error)
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name RoomRepository
