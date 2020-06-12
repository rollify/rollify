package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage/memory"
)

func TestDiceRollRepositoryCreateDiceRoll(t *testing.T) {
	tests := map[string]struct {
		roomID      string
		userID      string
		repo        func() *memory.DiceRollRepository
		diceRoll    model.DiceRoll
		expDiceRoll model.DiceRoll
		expErr      error
	}{

		"Having a dice roll without room ID should return a not valid error.": {
			roomID: "room-id",
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{ID: "test-id"},
			expErr:   internalerrors.ErrNotValid,
		},

		"Having a dice roll without user ID should return a not valid error.": {
			userID: "user-id",
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{ID: "test-id"},
			expErr:   internalerrors.ErrNotValid,
		},

		"Having a dice roll without ID should return a not valid error.": {
			roomID: "room-id",
			userID: "user-id",
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID: "",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Creating a dice roll that already exists should return an error.": {
			roomID: "room-id",
			userID: "user-id",
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
				ID: "test-id",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a dice roll should store the room.": {
			roomID: "room-id",
			userID: "user-id",
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID: "test-id",
			},
			expDiceRoll: model.DiceRoll{
				ID: "test-id",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			err := r.CreateDiceRoll(context.TODO(), test.roomID, test.userID, test.diceRoll)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the dice roll has been created internally.
				// This is very tide to the implementation but we need to access
				// internal structure to see if the implementation actually works.
				gotDiceRoll := r.DiceRollsByID[test.expDiceRoll.ID]
				assert.Equal(test.expDiceRoll, *gotDiceRoll)

				gotDiceRoll = r.DiceRollsByRoom[test.roomID]
				assert.Equal(test.expDiceRoll, *gotDiceRoll)

				gotDiceRoll = r.DiceRollsByRoomAndUser[test.roomID+test.userID]
				assert.Equal(test.expDiceRoll, *gotDiceRoll)
			}
		})
	}
}
