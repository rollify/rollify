package memory

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
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

type cursor struct {
	Serial int `json:"serial"`
}

// ListDiceRolls satisfies storage.DiceRollRepository interface.
func (r *DiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts storage.ListDiceRollsOpts) (*storage.DiceRollList, error) {
	if filterOpts.RoomID == "" {
		return nil, internalerrors.ErrNotValid
	}

	// Filter.
	var items []*model.DiceRoll
	// If no user means all room.
	if filterOpts.UserID == "" {
		items = r.DiceRollsByRoom[filterOpts.RoomID]
	} else {
		items = r.DiceRollsByRoomAndUser[filterOpts.RoomID+filterOpts.UserID]
	}

	// Get serial from cursor.
	// We use -1 because to search the starting point of the cursor, is the next one of the received serial,
	// if we don't receive cursor, that would be 0 cursor, and the next one is 1, so we would loose the first
	// user.
	userCursor := cursor{Serial: -1}
	if pageOpts.Cursor != "" {
		c, err := base64.StdEncoding.DecodeString(pageOpts.Cursor)
		if err != nil {
			return nil, fmt.Errorf("could not decode base64 cursor: %w", err)
		}
		err = json.Unmarshal(c, &userCursor)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal json cursor: %w", err)
		}
	}

	// Sort and get starting point based on cursor.
	// By default uses desc order (newest first).
	startIndex := 0
	found := false
	if pageOpts.Order == model.PaginationOrderAsc {
		sort.SliceStable(items, func(i, j int) bool { return items[i].Serial < items[j].Serial })
		// Search for the starting point of the list based on the cursor.
		// (Not optimized, don't need).
		for i, dr := range items {
			if int(dr.Serial) > userCursor.Serial {
				startIndex = i
				found = true
				break
			}
		}
	} else {
		sort.SliceStable(items, func(i, j int) bool { return items[i].Serial > items[j].Serial })
		// Filter (Not optimized, don't need).
		for i, dr := range items {
			if int(dr.Serial) < userCursor.Serial {
				startIndex = i
				found = true
				break
			}
		}
	}

	// If we have cursor and we didn't found, we are at the end of the list.
	if pageOpts.Cursor != "" && !found {
		startIndex = len(items)
	}

	// Filter the list from the starting point that we got with the cursor and
	// then cut wiht the size limit.
	filteredItems := items[startIndex:]
	hasNext := false
	if len(filteredItems) > int(pageOpts.Size) {
		hasNext = true
		filteredItems = filteredItems[:pageOpts.Size]
	}

	// Create our cursors.
	firstCursor := ""
	lastCursor := ""
	if len(filteredItems) > 0 {
		c := cursor{Serial: int(filteredItems[0].Serial)}
		jc, err := json.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("could not marshal cursor: %w", err)
		}
		firstCursor = base64.StdEncoding.EncodeToString([]byte(jc))

		c = cursor{Serial: int(filteredItems[len(filteredItems)-1].Serial)}
		jc, err = json.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("could not marshal cursor: %w", err)
		}
		lastCursor = base64.StdEncoding.EncodeToString([]byte(jc))
	}

	resultItems := make([]model.DiceRoll, 0, len(items))
	for _, v := range filteredItems {
		resultItems = append(resultItems, *v)
	}

	return &storage.DiceRollList{
		Items: resultItems,
		Cursors: model.PaginationCursors{
			FirstCursor: firstCursor,
			LastCursor:  lastCursor,
			HasPrevious: startIndex != 0, // If we are not the first means that we have previous.
			HasNext:     hasNext,
		},
	}, nil
}

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
