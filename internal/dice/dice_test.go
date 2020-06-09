package dice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice"
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
