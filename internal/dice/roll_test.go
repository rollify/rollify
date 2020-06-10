package dice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
)

func TestRandomRoller(t *testing.T) {
	tests := map[string]struct {
		diceRoll func() model.DiceRoll
		expErr   bool
	}{
		"Having a single die it should roll randomly": {
			diceRoll: func() model.DiceRoll {
				return model.DiceRoll{
					Dice: []model.DieRoll{
						{Type: model.DieTypeD6},
					},
				}
			},
		},
		"Having a multiple dice of the same type it should roll randomly": {
			diceRoll: func() model.DiceRoll {
				return model.DiceRoll{
					Dice: []model.DieRoll{
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
					},
				}
			},
		},

		"Having a multiple dice of different types it should roll randomly": {
			diceRoll: func() model.DiceRoll {
				return model.DiceRoll{
					Dice: []model.DieRoll{
						{Type: model.DieTypeD4, Side: 99999},
						{Type: model.DieTypeD6, Side: 99999},
						{Type: model.DieTypeD8, Side: 99999},
						{Type: model.DieTypeD10, Side: 99999},
						{Type: model.DieTypeD12, Side: 99999},
						{Type: model.DieTypeD20, Side: 99999},
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := dice.NewRandomRoller()

			gotDice := test.diceRoll()
			err := r.Roll(context.TODO(), &gotDice)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				// Check roll values are valid sides.
				for _, die := range gotDice.Dice {
					assert.GreaterOrEqual(die.Side, uint(0))
					assert.LessOrEqual(die.Side, die.Type.Sides()-1)
				}
			}
		})
	}
}
