package storage

import (
	"context"
	"time"

	"github.com/rollify/rollify/internal/model"
)

// DiceRollRepositoryMetricsRecorder knows how to measure DiceRollRepository.
type DiceRollRepositoryMetricsRecorder interface {
	MeasureDiceRollRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name DiceRollRepositoryMetricsRecorder

type measuredDiceRollRepository struct {
	storageType string
	rec         DiceRollRepositoryMetricsRecorder
	next        DiceRollRepository
}

// NewMeasuredDiceRollRepository wraps a DiceRollRepository and measures.
func NewMeasuredDiceRollRepository(storageType string, rec DiceRollRepositoryMetricsRecorder, next DiceRollRepository) DiceRollRepository {
	return &measuredDiceRollRepository{
		storageType: storageType,
		rec:         rec,
		next:        next,
	}
}

func (m measuredDiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceRollRepoOpDuration(ctx, m.storageType, "CreateDiceRoll", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateDiceRoll(ctx, dr)
}

func (m measuredDiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts ListDiceRollsOpts) (resp *DiceRollList, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureDiceRollRepoOpDuration(ctx, m.storageType, "ListDiceRolls", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.ListDiceRolls(ctx, pageOpts, filterOpts)
}

// RoomRepositoryMetricsRecorder knows how to measure RoomRepository.
type RoomRepositoryMetricsRecorder interface {
	MeasureRoomRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name RoomRepositoryMetricsRecorder

type measuredRoomRepository struct {
	storageType string
	rec         RoomRepositoryMetricsRecorder
	next        RoomRepository
}

// NewMeasuredRoomRepository wraps a RoomRepository and measures.
func NewMeasuredRoomRepository(storageType string, rec RoomRepositoryMetricsRecorder, next RoomRepository) RoomRepository {
	return &measuredRoomRepository{
		storageType: storageType,
		rec:         rec,
		next:        next,
	}
}

func (m measuredRoomRepository) CreateRoom(ctx context.Context, r model.Room) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureRoomRepoOpDuration(ctx, m.storageType, "CreateRoom", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateRoom(ctx, r)
}

func (m measuredRoomRepository) GetRoom(ctx context.Context, id string) (room *model.Room, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureRoomRepoOpDuration(ctx, m.storageType, "GetRoom", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.GetRoom(ctx, id)
}

func (m measuredRoomRepository) RoomExists(ctx context.Context, id string) (exists bool, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureRoomRepoOpDuration(ctx, m.storageType, "RoomExists", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.RoomExists(ctx, id)
}

// UserRepositoryMetricsRecorder knows how to measure UserRepository.
type UserRepositoryMetricsRecorder interface {
	MeasureUserRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration)
}

//go:generate mockery --case underscore --output storagemock --outpkg storagemock --name UserRepositoryMetricsRecorder

type measuredUserRepository struct {
	storageType string
	rec         UserRepositoryMetricsRecorder
	next        UserRepository
}

// NewMeasuredUserRepository wraps a UserRepository and measures.
func NewMeasuredUserRepository(storageType string, rec UserRepositoryMetricsRecorder, next UserRepository) UserRepository {
	return &measuredUserRepository{
		storageType: storageType,
		rec:         rec,
		next:        next,
	}
}

func (m measuredUserRepository) CreateUser(ctx context.Context, u model.User) (err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "CreateUser", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.CreateUser(ctx, u)
}

func (m measuredUserRepository) ListRoomUsers(ctx context.Context, roomID string) (ul *UserList, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "ListRoomUsers", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.ListRoomUsers(ctx, roomID)
}

func (m measuredUserRepository) GetUserByID(ctx context.Context, userID string) (u *model.User, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "GetUserByID", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.GetUserByID(ctx, userID)
}

func (m measuredUserRepository) UserExists(ctx context.Context, userID string) (ex bool, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "UserExists", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.UserExists(ctx, userID)
}

func (m measuredUserRepository) UserExistsByNameInsensitive(ctx context.Context, roomID, username string) (ex bool, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "UserExistsByNameInsensitive", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.UserExistsByNameInsensitive(ctx, roomID, username)
}

func (m measuredUserRepository) GetUserByNameInsensitive(ctx context.Context, roomID, username string) (u *model.User, err error) {
	defer func(t0 time.Time) {
		m.rec.MeasureUserRepoOpDuration(ctx, m.storageType, "GetUserByNameInsensitive", err == nil, time.Since(t0))
	}(time.Now())

	return m.next.GetUserByNameInsensitive(ctx, roomID, username)
}
