package user

import (
	"context"
	"time"
)

// ServiceMetricsRecorder knows how to record Service metrics.
type ServiceMetricsRecorder interface {
	MeasureUserServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output usermock --outpkg usermock --name ServiceMetricsRecorder

type measuredService struct {
	rec  ServiceMetricsRecorder
	next Service
}

// NewMeasureService wraps a service and measures.
func NewMeasureService(rec ServiceMetricsRecorder, next Service) Service {
	return &measuredService{
		rec:  rec,
		next: next,
	}
}

func (m measuredService) CreateUser(ctx context.Context, req CreateUserRequest) (resp *CreateUserResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserServiceOpDuration(ctx, "CreateUser", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateUser(ctx, req)
}

func (m measuredService) ListUsers(ctx context.Context, req ListUsersRequest) (resp *ListUsersResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserServiceOpDuration(ctx, "ListUsers", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.ListUsers(ctx, req)
}

func (m measuredService) GetUser(ctx context.Context, req GetUserRequest) (resp *GetUserResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserServiceOpDuration(ctx, "GetUser", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.GetUser(ctx, req)
}
