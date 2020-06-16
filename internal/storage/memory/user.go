package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// UserRepository is a fake repository based on memory.
// This repository exposes the storage to the public so the users can
// check the internal data in and maniputale it (e.g tests)
type UserRepository struct {
	// UsersByRoom is where the users data is stored by room. Not thread safe.
	UsersByRoom map[string]map[string]*model.User

	mu sync.Mutex
}

// NewUserRepository returns a new UserRepository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		UsersByRoom: map[string]map[string]*model.User{},
	}
}

// CreateUser satisfies storage.UserRepository interface.
func (r *UserRepository) CreateUser(ctx context.Context, u model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch {
	case u.ID == "":
		return fmt.Errorf("missing ID: %w", internalerrors.ErrNotValid)
	case u.RoomID == "":
		return fmt.Errorf("missing RoomID: %w", internalerrors.ErrNotValid)
	case u.Name == "":
		return fmt.Errorf("missing Name: %w", internalerrors.ErrNotValid)
	}

	_, ok := r.UsersByRoom[u.RoomID]
	if !ok {
		r.UsersByRoom[u.RoomID] = map[string]*model.User{}
	}

	_, ok = r.UsersByRoom[u.RoomID][u.ID]
	if ok {
		return fmt.Errorf("user already exists: %w", internalerrors.ErrAlreadyExists)
	}

	r.UsersByRoom[u.RoomID][u.ID] = &u

	return nil
}

// ListRoomUsers satisfies storage.UserRepository interface.
func (r *UserRepository) ListRoomUsers(ctx context.Context, roomID string) (*storage.UserList, error) {
	if roomID == "" {
		return nil, fmt.Errorf("missing RoomID: %w", internalerrors.ErrNotValid)
	}

	us := r.UsersByRoom[roomID]

	users := make([]model.User, 0, len(us))
	for _, u := range us {
		users = append(users, *u)
	}

	return &storage.UserList{
		Items: users,
	}, nil
}

// Implementation assertions.
var _ storage.UserRepository = &UserRepository{}
