// Code generated by mockery v2.0.0-alpha.2. DO NOT EDIT.

package eventmock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/rollify/rollify/internal/model"
)

// Notifier is an autogenerated mock type for the Notifier type
type Notifier struct {
	mock.Mock
}

// NotifyDiceRollCreated provides a mock function with given fields: ctx, e
func (_m *Notifier) NotifyDiceRollCreated(ctx context.Context, e model.EventDiceRollCreated) error {
	ret := _m.Called(ctx, e)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.EventDiceRollCreated) error); ok {
		r0 = rf(ctx, e)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
