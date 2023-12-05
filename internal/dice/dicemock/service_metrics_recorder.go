// Code generated by mockery v2.33.1. DO NOT EDIT.

package dicemock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// ServiceMetricsRecorder is an autogenerated mock type for the ServiceMetricsRecorder type
type ServiceMetricsRecorder struct {
	mock.Mock
}

// MeasureDiceServiceOpDuration provides a mock function with given fields: ctx, op, success, t
func (_m *ServiceMetricsRecorder) MeasureDiceServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration) {
	_m.Called(ctx, op, success, t)
}

// NewServiceMetricsRecorder creates a new instance of ServiceMetricsRecorder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewServiceMetricsRecorder(t interface {
	mock.TestingT
	Cleanup(func())
}) *ServiceMetricsRecorder {
	mock := &ServiceMetricsRecorder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
