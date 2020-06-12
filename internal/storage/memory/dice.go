package memory

import (
	"context"
	"sync"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// DiceRollRepository is a fake repository based on memory.
// This repository exposes the storage to the public so the users can
// check the internal data in and maniputale it (e.g tests)
type DiceRollRepository struct {
	// DiceRollsByID is where the dice roll data is stored by ID. Not thread safe.
	DiceRollsByID map[string]*model.DiceRoll
	// DiceRollsByRoom is where the dice roll data is stored by room. Not thread safe.
	DiceRollsByRoom map[string]*model.DiceRoll
	// DiceRollsByRoomAndUser is where the dice roll data is stored by room and user. Not thread safe.
	DiceRollsByRoomAndUser map[string]*model.DiceRoll

	mu sync.Mutex
}

// NewDiceRollRepository returns a new DiceRollRepository.
func NewDiceRollRepository() *DiceRollRepository {
	return &DiceRollRepository{
		DiceRollsByID:          map[string]*model.DiceRoll{},
		DiceRollsByRoom:        map[string]*model.DiceRoll{},
		DiceRollsByRoomAndUser: map[string]*model.DiceRoll{},
	}
}

// CreateDiceRoll satisfies dice.Repository interface.
func (r *DiceRollRepository) CreateDiceRoll(ctx context.Context, roomID, userID string, dr model.DiceRoll) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if roomID == "" || userID == "" || dr.ID == "" {
		return internalerrors.ErrNotValid
	}

	_, ok := r.DiceRollsByID[dr.ID]
	if ok {
		return internalerrors.ErrAlreadyExists
	}

	r.DiceRollsByID[dr.ID] = &dr
	r.DiceRollsByRoom[roomID] = &dr
	r.DiceRollsByRoomAndUser[roomID+userID] = &dr

	return nil
}

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
