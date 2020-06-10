package dice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/model"
)

func TestServiceListDiceTypes(t *testing.T) {
	tests := map[string]struct {
		config  dice.ServiceConfig
		expResp func() *dice.ListDiceTypesResponse
		expErr  bool
	}{
		"Listing dice types should return all the available dice types.": {
			expResp: func() *dice.ListDiceTypesResponse {
				return &dice.ListDiceTypesResponse{
					DiceTypes: []model.DieType{
						model.DieTypeD4,
						model.DieTypeD6,
						model.DieTypeD8,
						model.DieTypeD10,
						model.DieTypeD12,
						model.DieTypeD20,
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			test.config.Roller = &dicemock.Roller{}
			svc, err := dice.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.ListDiceTypes(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceRollDice(t *testing.T) {
	tests := map[string]struct {
		config  dice.ServiceConfig
		mock    func(r *dicemock.Roller)
		req     func() dice.CreateDiceRollRequest
		expResp func() *dice.CreateDiceRollResponse
		expErr  bool
	}{
		"Having a dice roll request without room should fail.": {
			mock: func(r *dicemock.Roller) {},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request without user should fail.": {
			mock: func(r *dicemock.Roller) {},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "room-id",
					UserID: "",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request without dice should fail.": {
			mock: func(r *dicemock.Roller) {},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "room-id",
					UserID: "test-id",
					Dice:   []model.DieType{},
				}
			},
			expErr: true,
		},

		"Having a dice roll request it should create a dice roll and roll them.": {
			mock: func(r *dicemock.Roller) {
				// Expexted dice roll call.
				exp := &model.DiceRoll{
					ID: "test",
					Dice: []model.DieRoll{
						{ID: "test", Type: model.DieTypeD6},
						{ID: "test", Type: model.DieTypeD8},
						{ID: "test", Type: model.DieTypeD10},
					},
				}
				r.On("Roll", mock.Anything, exp).Once().Return(nil)
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice: []model.DieType{
						model.DieTypeD6,
						model.DieTypeD8,
						model.DieTypeD10,
					},
				}
			},
			expResp: func() *dice.CreateDiceRollResponse {
				return &dice.CreateDiceRollResponse{
					DiceRoll: model.DiceRoll{
						ID: "test",
						Dice: []model.DieRoll{
							{ID: "test", Type: model.DieTypeD6},
							{ID: "test", Type: model.DieTypeD8},
							{ID: "test", Type: model.DieTypeD10},
						},
					},
				}
			},
		},

		"Having a dice roll request and failing the dice roll process, it should fail..": {
			mock: func(r *dicemock.Roller) {
				r.On("Roll", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("wanted error"))
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mr := &dicemock.Roller{}
			test.mock(mr)

			test.config.Roller = mr
			test.config.IDGenerator = func() string { return "test" }

			svc, err := dice.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.CreateDiceRoll(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}
