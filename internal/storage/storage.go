package storage

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// DiceRollList is a list of dice rolls.
type DiceRollList struct {
	Items   []model.DiceRoll
	Cursors model.PaginationCursors
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
	// If the dice data is missing or not valid it will return a internalerrors.NotValid error kind.
	// If the dice roll already exists it returns a internalerrors.AlreadyExists error kind.
	CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error
	// ListDiceRolls lists dice rolls, by default in descendant order (newest first).
	// If the dice roomID option is empty it returns a internalerrors.NotValid error kind.
	ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts ListDiceRollsOpts) (*DiceRollList, error)
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name DiceRollRepository

// RoomRepository is the repository interface that implementations need to
// implement to manage rooms in storage.
type RoomRepository interface {
	// CreateRoom creates a new room.
	// If the room data is missing or not valid it will return a internalerrors.NotValid error kind.
	// If the room already exists it returns a internalerrors.AlreadyExists error kind.
	CreateRoom(ctx context.Context, r model.Room) error
	// RoomExists returns true if the room exists.
	RoomExists(ctx context.Context, id string) (exists bool, err error)
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name RoomRepository

// UserList is a list of users.
type UserList struct {
	Items []model.User
}

// UserRepository is the repository interface that implementations need to
// implement to manage users in storage.
type UserRepository interface {
	// CreateRoom creates a new room.
	// If the user data is missing or not valid it will return a internalerrors.NotValid error kind.
	// If the room already exists it returns a internalerrors.AlreadyExists error kind.
	CreateUser(ctx context.Context, u model.User) error
	// ListRoomUsers returns the user list of a room.
	ListRoomUsers(ctx context.Context, roomID string) (*UserList, error)
	// UserExists returns true if the ID of the user exists.
	UserExists(ctx context.Context, userID string) (bool, error)
	// UserExistsByNameInsensitive checks if a user exists in a room using the username
	// in case insensitive mode.
	UserExistsByNameInsensitive(ctx context.Context, roomID, username string) (bool, error)
}

//go:generate mockery -case underscore -output storagemock -outpkg storagemock -name UserRepository
