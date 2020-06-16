// Code generated by mockery v2.0.0-alpha.2. DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/rollify/rollify/internal/model"
	mock "github.com/stretchr/testify/mock"

	storage "github.com/rollify/rollify/internal/storage"
)

// UserRepository is an autogenerated mock type for the UserRepository type
type UserRepository struct {
	mock.Mock
}

// CreateUser provides a mock function with given fields: ctx, u
func (_m *UserRepository) CreateUser(ctx context.Context, u model.User) error {
	ret := _m.Called(ctx, u)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.User) error); ok {
		r0 = rf(ctx, u)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListRoomUsers provides a mock function with given fields: ctx, roomID
func (_m *UserRepository) ListRoomUsers(ctx context.Context, roomID string) (*storage.UserList, error) {
	ret := _m.Called(ctx, roomID)

	var r0 *storage.UserList
	if rf, ok := ret.Get(0).(func(context.Context, string) *storage.UserList); ok {
		r0 = rf(ctx, roomID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.UserList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, roomID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}