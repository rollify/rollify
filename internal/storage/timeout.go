package storage

import (
	"context"
	"time"

	"github.com/rollify/rollify/internal/model"
)

type timeoutDiceRollRepository struct {
	timeout time.Duration
	next    DiceRollRepository
}

// NewTimeoutDiceRollRepository wraps a DiceRollRepository and timeouts.
func NewTimeoutDiceRollRepository(timeout time.Duration, next DiceRollRepository) DiceRollRepository {
	return &timeoutDiceRollRepository{
		timeout: timeout,
		next:    next,
	}
}

func (t timeoutDiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) (err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.CreateDiceRoll(ctx, dr)
}

func (t timeoutDiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts ListDiceRollsOpts) (resp *DiceRollList, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.ListDiceRolls(ctx, pageOpts, filterOpts)
}

type timeoutRoomRepository struct {
	timeout time.Duration
	next    RoomRepository
}

// NewTimeoutRoomRepository wraps a RoomRepository and timeouts.
func NewTimeoutRoomRepository(timeout time.Duration, next RoomRepository) RoomRepository {
	return &timeoutRoomRepository{
		timeout: timeout,
		next:    next,
	}
}

func (t timeoutRoomRepository) CreateRoom(ctx context.Context, r model.Room) (err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.CreateRoom(ctx, r)
}

func (t timeoutRoomRepository) GetRoom(ctx context.Context, id string) (room *model.Room, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.GetRoom(ctx, id)
}

func (t timeoutRoomRepository) RoomExists(ctx context.Context, id string) (exists bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.RoomExists(ctx, id)
}

type timeoutUserRepository struct {
	timeout time.Duration
	next    UserRepository
}

// NewTimeoutUserRepository wraps a UserRepository and timeouts.
func NewTimeoutUserRepository(timeout time.Duration, next UserRepository) UserRepository {
	return &timeoutUserRepository{
		timeout: timeout,
		next:    next,
	}
}

func (t timeoutUserRepository) CreateUser(ctx context.Context, u model.User) (err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.CreateUser(ctx, u)
}

func (t timeoutUserRepository) ListRoomUsers(ctx context.Context, roomID string) (ul *UserList, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.ListRoomUsers(ctx, roomID)
}

func (t timeoutUserRepository) UserExists(ctx context.Context, userID string) (ex bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.UserExists(ctx, userID)
}

func (t timeoutUserRepository) UserExistsByNameInsensitive(ctx context.Context, roomID, username string) (ex bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.next.UserExistsByNameInsensitive(ctx, roomID, username)
}
