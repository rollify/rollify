package room

import (
	"context"

	"github.com/rollify/rollify/internal/model"
)

// Repository is the repository interface that implementations need to
// implement to manage rooms in storage.
type Repository interface {
	// CreateRoom creates a new room.
	// If the room doesn't have an ID it returns a internalerrors.NotValid error kind.
	// If the room already exists it returns a internalerrors.AlreadyExists error kind.
	CreateRoom(ctx context.Context, r model.Room) error
}

//go:generate mockery -case underscore -output roommock -outpkg roommock -name Repository
