// Code generated by mockery v1.0.0. DO NOT EDIT.

package dicemock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/rollify/rollify/internal/model"
)

// Roller is an autogenerated mock type for the Roller type
type Roller struct {
	mock.Mock
}

// Roll provides a mock function with given fields: ctx, d
func (_m *Roller) Roll(ctx context.Context, d *model.DiceRoll) error {
	ret := _m.Called(ctx, d)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.DiceRoll) error); ok {
		r0 = rf(ctx, d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
