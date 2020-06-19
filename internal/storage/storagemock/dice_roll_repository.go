// Code generated by mockery v2.0.0-alpha.2. DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/rollify/rollify/internal/model"
	mock "github.com/stretchr/testify/mock"

	storage "github.com/rollify/rollify/internal/storage"
)

// DiceRollRepository is an autogenerated mock type for the DiceRollRepository type
type DiceRollRepository struct {
	mock.Mock
}

// CreateDiceRoll provides a mock function with given fields: ctx, dr
func (_m *DiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error {
	ret := _m.Called(ctx, dr)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.DiceRoll) error); ok {
		r0 = rf(ctx, dr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListDiceRolls provides a mock function with given fields: ctx, pageOpts, filterOpts
func (_m *DiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts storage.PaginationOpts, filterOpts storage.ListDiceRollsOpts) (*storage.DiceRollList, error) {
	ret := _m.Called(ctx, pageOpts, filterOpts)

	var r0 *storage.DiceRollList
	if rf, ok := ret.Get(0).(func(context.Context, storage.PaginationOpts, storage.ListDiceRollsOpts) *storage.DiceRollList); ok {
		r0 = rf(ctx, pageOpts, filterOpts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.DiceRollList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, storage.PaginationOpts, storage.ListDiceRollsOpts) error); ok {
		r1 = rf(ctx, pageOpts, filterOpts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
