package memory

import (
	"context"
	"fmt"
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
	DiceRollsByRoom map[string][]*model.DiceRoll
	// DiceRollsByRoomAndUser is where the dice roll data is stored by room and user. Not thread safe.
	DiceRollsByRoomAndUser map[string][]*model.DiceRoll

	// serialTrack will track the point where the serials for diceRolls are.
	serialTrack uint
	mu          sync.Mutex
}

// NewDiceRollRepository returns a new DiceRollRepository.
func NewDiceRollRepository() *DiceRollRepository {
	return &DiceRollRepository{
		DiceRollsByID:          map[string]*model.DiceRoll{},
		DiceRollsByRoom:        map[string][]*model.DiceRoll{},
		DiceRollsByRoomAndUser: map[string][]*model.DiceRoll{},
	}
}

// CreateDiceRoll satisfies dice.Repository interface.
func (r *DiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch {
	case dr.RoomID == "":
		return fmt.Errorf("missing room ID: %w", internalerrors.ErrNotValid)
	case dr.UserID == "":
		return fmt.Errorf("missing user ID: %w", internalerrors.ErrNotValid)
	case dr.ID == "":
		return fmt.Errorf("missing ID: %w", internalerrors.ErrNotValid)
	}

	_, ok := r.DiceRollsByID[dr.ID]
	if ok {
		return internalerrors.ErrAlreadyExists
	}

	r.DiceRollsByID[dr.ID] = &dr
	r.DiceRollsByRoom[dr.RoomID] = append(r.DiceRollsByRoom[dr.RoomID], &dr)
	r.DiceRollsByRoomAndUser[dr.RoomID+dr.UserID] = append(r.DiceRollsByRoomAndUser[dr.RoomID+dr.UserID], &dr)

	// Set up the serial.
	dr.Serial = r.serialTrack
	r.serialTrack++

	return nil
}

// ListDiceRolls satisfies storage.DiceRollRepository interface.
func (r *DiceRollRepository) ListDiceRolls(ctx context.Context, opts storage.ListDiceRollsOpts) (*storage.DiceRollList, error) {
	if opts.RoomID == "" {
		return nil, internalerrors.ErrNotValid
	}

	var items []*model.DiceRoll

	// If no user means all room.
	if opts.UserID == "" {
		items = r.DiceRollsByRoom[opts.RoomID]
	} else {
		items = r.DiceRollsByRoomAndUser[opts.RoomID+opts.UserID]
	}

	resultItems := make([]model.DiceRoll, 0, len(items))
	for _, v := range items {
		resultItems = append(resultItems, *v)
	}

	return &storage.DiceRollList{
		Items: resultItems,
	}, nil
}

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
