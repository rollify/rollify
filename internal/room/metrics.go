package room

import (
	"context"
	"time"
)

// ServiceMetricsRecorder knows how to record Service metrics.
type ServiceMetricsRecorder interface {
	MeasureRoomServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output roommock --outpkg roommock --name ServiceMetricsRecorder

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

func (m measuredService) CreateRoom(ctx context.Context, req CreateRoomRequest) (resp *CreateRoomResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureRoomServiceOpDuration(ctx, "CreateRoom", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateRoom(ctx, req)
}

func (m measuredService) GetRoom(ctx context.Context, req GetRoomRequest) (resp *GetRoomResponse, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureRoomServiceOpDuration(ctx, "GetRoom", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.GetRoom(ctx, req)
}
