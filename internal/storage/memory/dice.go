package memory

import (
	"context"
	"sync"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// DiceRollRepository is a memory based dice rolls repository.
type DiceRollRepository struct {
	mu            sync.Mutex
	diceRollsByID map[string]model.DiceRoll
}

// NewDiceRollRepository returns a new DiceRollRepository.
func NewDiceRollRepository() *DiceRollRepository {
	return &DiceRollRepository{
		diceRollsByID: map[string]model.DiceRoll{},
	}
}

// SetDiceRollsByIDSeed helper function to set the data we want at any point.
func (r *DiceRollRepository) SetDiceRollsByIDSeed(data map[string]model.DiceRoll) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.diceRollsByID = data
}

// DiceRollsByIDSeed helper function to get the data we want at any point.
func (r *DiceRollRepository) DiceRollsByIDSeed() map[string]model.DiceRoll {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.diceRollsByID
}

// CreateDiceRoll satisfies dice.Repository interface.
func (r *DiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if dr.ID == "" {
		return internalerrors.ErrNotValid
	}

	_, ok := r.diceRollsByID[dr.ID]
	if ok {
		return internalerrors.ErrAlreadyExists
	}

	r.diceRollsByID[dr.ID] = dr

	return nil
}

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
