package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/memory"
)

func TestDiceRollRepositoryCreateDiceRoll(t *testing.T) {
	tests := map[string]struct {
		repo        func() *memory.DiceRollRepository
		diceRoll    model.DiceRoll
		expDiceRoll model.DiceRoll
		expErr      error
	}{

		"Having a dice roll without room ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a dice roll without user ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Having a dice roll without ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Creating a dice roll that already exists should return an error.": {
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByID = map[string]*model.DiceRoll{
					"test-id": {
						ID: "test-id",
					},
				}
				return r
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a dice roll should store the room.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
			expDiceRoll: model.DiceRoll{
				ID:     "test-id",
				RoomID: "room-id",
				UserID: "user-id",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			err := r.CreateDiceRoll(context.TODO(), test.diceRoll)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the dice roll has been created internally.
				// This is very tide to the implementation but we need to access
				// internal structure to see if the implementation actually works.
				gotDiceRoll := r.DiceRollsByID[test.expDiceRoll.ID]
				assert.Equal(test.expDiceRoll, *gotDiceRoll)

				gotDiceRolls := r.DiceRollsByRoom[test.diceRoll.RoomID]
				assert.Contains(gotDiceRolls, &test.expDiceRoll)

				gotDiceRolls = r.DiceRollsByRoomAndUser[test.diceRoll.RoomID+test.diceRoll.UserID]
				assert.Contains(gotDiceRolls, &test.expDiceRoll)
			}
		})
	}
}

func TestDiceRollRepositoryListDiceRoll(t *testing.T) {
	tests := map[string]struct {
		opts         storage.ListDiceRollsOpts
		repo         func() *memory.DiceRollRepository
		expDiceRolls *storage.DiceRollList
		expErr       error
	}{
		"Listing dice rolls without room should fail.": {
			opts: storage.ListDiceRollsOpts{
				RoomID: "",
				UserID: "",
			},
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Listing all users dice rolls should return them.": {
			opts: storage.ListDiceRollsOpts{
				RoomID: "room-1",
				UserID: "",
			},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoom = map[string][]*model.DiceRoll{
					"room-0": {{ID: "test00"}},
					"room-1": {{ID: "test10"}, {ID: "test11"}},
					"room-2": {{ID: "test20"}},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test10"},
					{ID: "test11"},
				},
			},
		},

		"Listing single user dice rolls in a room should return them.": {
			opts: storage.ListDiceRollsOpts{
				RoomID: "room-2",
				UserID: "user-1",
			},
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.DiceRollsByRoomAndUser = map[string][]*model.DiceRoll{
					"room-0user-1": {{ID: "test00"}},
					"room-1user-2": {{ID: "test10"}, {ID: "test11"}},
					"room-2user-1": {{ID: "test20"}},
				}
				return r
			},
			expDiceRolls: &storage.DiceRollList{
				Items: []model.DiceRoll{
					{ID: "test20"},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotDiceRolls, err := r.ListDiceRolls(context.TODO(), test.opts)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				assert.Equal(test.expDiceRolls, gotDiceRolls)
			}
		})
	}
}
