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
		repo        func() *memory.DiceRollRepository
		diceRoll    model.DiceRoll
		expDiceRoll model.DiceRoll
		expErr      error
	}{
		"Having a dice roll without ID should return a not valid error.": {
			repo: func() *memory.DiceRollRepository {
				return memory.NewDiceRollRepository()
			},
			diceRoll: model.DiceRoll{
				ID: "",
			},
			expErr: internalerrors.ErrNotValid,
		},

		"Creating a dice roll that already exists should return an error.": {
			repo: func() *memory.DiceRollRepository {
				r := memory.NewDiceRollRepository()
				r.SetDiceRollsByIDSeed(map[string]model.DiceRoll{
					"test-id": {
						ID: "test-id",
					},
				})
				return r
			},
			diceRoll: model.DiceRoll{
				ID: "test-id",
			},
			expErr: internalerrors.ErrAlreadyExists,
		},

		"Creating a dice roll should store the room.": {
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
			err := r.CreateDiceRoll(context.TODO(), test.diceRoll)

			if test.expErr != nil && assert.Error(err) {
				assert.True(errors.Is(err, test.expErr))
			} else if assert.NoError(err) {
				// Check the dice roll has been created internally.
				seed := r.DiceRollsByIDSeed()
				gotDiceRoll := seed[test.expDiceRoll.ID]
				assert.Equal(test.expDiceRoll, gotDiceRoll)
			}
		})
	}
}
