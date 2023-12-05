// Code generated by mockery v2.33.1. DO NOT EDIT.

package eventmock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/rollify/rollify/internal/model"
)

// Subscriber is an autogenerated mock type for the Subscriber type
type Subscriber struct {
	mock.Mock
}

// SubscribeDiceRollCreated provides a mock function with given fields: ctx, subscribeID, roomID, h
func (_m *Subscriber) SubscribeDiceRollCreated(ctx context.Context, subscribeID string, roomID string, h func(context.Context, model.EventDiceRollCreated) error) error {
	ret := _m.Called(ctx, subscribeID, roomID, h)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, func(context.Context, model.EventDiceRollCreated) error) error); ok {
		r0 = rf(ctx, subscribeID, roomID, h)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubscribeDiceRollCreated provides a mock function with given fields: ctx, subscribeID, roomID
func (_m *Subscriber) UnsubscribeDiceRollCreated(ctx context.Context, subscribeID string, roomID string) error {
	ret := _m.Called(ctx, subscribeID, roomID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, subscribeID, roomID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSubscriber creates a new instance of Subscriber. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSubscriber(t interface {
	mock.TestingT
	Cleanup(func())
}) *Subscriber {
	mock := &Subscriber{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
