// Code generated by mockery v2.0.0-alpha.2. DO NOT EDIT.

package dicemock

import (
	context "context"

	dice "github.com/rollify/rollify/internal/dice"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// CreateDiceRoll provides a mock function with given fields: ctx, r
func (_m *Service) CreateDiceRoll(ctx context.Context, r dice.CreateDiceRollRequest) (*dice.CreateDiceRollResponse, error) {
	ret := _m.Called(ctx, r)

	var r0 *dice.CreateDiceRollResponse
	if rf, ok := ret.Get(0).(func(context.Context, dice.CreateDiceRollRequest) *dice.CreateDiceRollResponse); ok {
		r0 = rf(ctx, r)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dice.CreateDiceRollResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, dice.CreateDiceRollRequest) error); ok {
		r1 = rf(ctx, r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDiceTypes provides a mock function with given fields: ctx
func (_m *Service) ListDiceTypes(ctx context.Context) (*dice.ListDiceTypesResponse, error) {
	ret := _m.Called(ctx)

	var r0 *dice.ListDiceTypesResponse
	if rf, ok := ret.Get(0).(func(context.Context) *dice.ListDiceTypesResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dice.ListDiceTypesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
