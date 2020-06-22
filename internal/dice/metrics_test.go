package dice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/model"
)

func TestMeasuredRollerRoll(t *testing.T) {
	tests := map[string]struct {
		rollerType string
		diceRoll   func() *model.DiceRoll
		mock       func(mrec *dicemock.RollerMetricsRecorder, mroll *dicemock.Roller)
		expErr     error
	}{
		"Havign a correct method call, should be measured and call the wrapped.": {
			rollerType: "test1",
			diceRoll: func() *model.DiceRoll {
				return &model.DiceRoll{
					Dice: []model.DieRoll{
						{Type: model.DieTypeD6, Side: 4},
						{Type: model.DieTypeD10, Side: 9},
						{Type: model.DieTypeD20, Side: 13},
					},
				}
			},
			mock: func(mrec *dicemock.RollerMetricsRecorder, mroll *dicemock.Roller) {
				expDie1 := model.DieRoll{Type: model.DieTypeD6, Side: 4}
				expDie2 := model.DieRoll{Type: model.DieTypeD10, Side: 9}
				expDie3 := model.DieRoll{Type: model.DieTypeD20, Side: 13}
				expDiceRoll := &model.DiceRoll{
					Dice: []model.DieRoll{expDie1, expDie2, expDie3},
				}

				mrec.On("MeasureDiceRollQuantity", mock.Anything, "test1", expDiceRoll)
				mrec.On("MeasureDieRollResult", mock.Anything, "test1", &expDie1)
				mrec.On("MeasureDieRollResult", mock.Anything, "test1", &expDie2)
				mrec.On("MeasureDieRollResult", mock.Anything, "test1", &expDie3)

				mroll.On("Roll", mock.Anything, expDiceRoll).Once().Return(nil)
			},
		},

		"Havign a call method that returns error, should return the error.": {
			rollerType: "test1",
			diceRoll: func() *model.DiceRoll {
				return &model.DiceRoll{}
			},
			mock: func(mrec *dicemock.RollerMetricsRecorder, mroll *dicemock.Roller) {
				mrec.On("MeasureDiceRollQuantity", mock.Anything, mock.Anything, mock.Anything).Maybe()
				mrec.On("MeasureDieRollResult", mock.Anything, mock.Anything, mock.Anything).Maybe()

				mroll.On("Roll", mock.Anything, mock.Anything).Once().Return(errors.New("wanted error"))
			},
			expErr: errors.New("wanted error"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mrec := &dicemock.RollerMetricsRecorder{}
			mroll := &dicemock.Roller{}
			test.mock(mrec, mroll)

			r := dice.NewMeasureRoller(test.rollerType, mrec, mroll)
			err := r.Roll(context.TODO(), test.diceRoll())

			mrec.AssertExpectations(t)
			mroll.AssertExpectations(t)
			assert.Equal(test.expErr, err)
		})
	}
}
