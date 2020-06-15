package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// RoomRepository is a fake repository based on memory.
// This repository exposes the storage to the public so the users can
// check the internal data in and maniputale it (e.g tests)
type RoomRepository struct {
	// RoomsByID is where the room data is stored by ID. Not thread safe.
	RoomsByID map[string]*model.Room

	mu sync.Mutex
}

// NewRoomRepository returns a new RoomRepository.
func NewRoomRepository() *RoomRepository {
	return &RoomRepository{
		RoomsByID: map[string]*model.Room{},
	}
}

// CreateRoom satisfies room.Repository interface.
func (r *RoomRepository) CreateRoom(_ context.Context, room model.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if room.ID == "" {
		return fmt.Errorf("missing room ID: %w", internalerrors.ErrNotValid)
	}

	_, ok := r.RoomsByID[room.ID]
	if ok {
		return internalerrors.ErrAlreadyExists
	}

	r.RoomsByID[room.ID] = &room

	return nil
}

// RoomExists satisfies room.Repository interface.
func (r *RoomRepository) RoomExists(_ context.Context, id string) (exists bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.RoomsByID[id]
	return ok, nil
}

// Implementation assertions.
var _ storage.RoomRepository = &RoomRepository{}
