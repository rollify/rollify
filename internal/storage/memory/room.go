package memory

import (
	"context"
	"sync"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// RoomRepository is a memory based room repository.
type RoomRepository struct {
	mu        sync.Mutex
	roomsByID map[string]model.Room
}

// NewRoomRepository returns a new RoomRepository.
func NewRoomRepository() *RoomRepository {
	return &RoomRepository{
		roomsByID: map[string]model.Room{},
	}
}

// SetRoomsByIDSeed helper function to set the data we want at any point.
func (r *RoomRepository) SetRoomsByIDSeed(data map[string]model.Room) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roomsByID = data
}

// RoomsByIDSeed helper function to get the data we want at any point.
func (r *RoomRepository) RoomsByIDSeed() map[string]model.Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.roomsByID
}

// CreateRoom satisfies room.Repository interface.
func (r *RoomRepository) CreateRoom(ctx context.Context, room model.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if room.ID == "" {
		return internalerrors.ErrNotValid
	}

	_, ok := r.roomsByID[room.ID]
	if ok {
		return internalerrors.ErrAlreadyExists
	}

	r.roomsByID[room.ID] = room

	return nil
}

// Implementation assertions.
var _ storage.RoomRepository = &RoomRepository{}
