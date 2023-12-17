package storage

import (
	"context"
	"fmt"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/rollify/rollify/internal/model"
)

type cachedRoomRepository struct {
	roomCache *lru.Cache[string, *model.Room]
	RoomRepository
}

// NewCachedRoomRepository wraps a RoomRepository and caches the rooms information in memory
// is not a cache to try to optimize the query to the original repository but try caching the
// information of the rooms that are asked frequently and save most of the room info accesses.
func NewCachedRoomRepository(next RoomRepository) (RoomRepository, error) {
	c, err := lru.New[string, *model.Room](500)
	if err != nil {
		return nil, fmt.Errorf("could not initialize cache")
	}

	return &cachedRoomRepository{
		roomCache:      c,
		RoomRepository: next,
	}, nil
}

func (c cachedRoomRepository) GetRoom(ctx context.Context, id string) (room *model.Room, err error) {
	r, ok := c.roomCache.Get(id)
	if ok {
		return r, nil
	}

	r, err = c.RoomRepository.GetRoom(ctx, id)
	if err != nil {
		return r, err
	}

	// Save in cache.
	_ = c.roomCache.Add(id, r)
	return r, err
}

func (c cachedRoomRepository) RoomExists(ctx context.Context, id string) (exists bool, err error) {
	// Try a best effort.
	_, ok := c.roomCache.Get(id)
	if ok {
		return true, nil
	}

	return c.RoomRepository.RoomExists(ctx, id)
}

type cachedUserRepository struct {
	userIDCache         *lru.Cache[string, *model.User]
	userNameExistsCache *lru.Cache[string, bool]
	userNameCache       *lru.Cache[string, *model.User]
	UserRepository
}

// NewCachedUserRepository wraps a UserRepository and caches the users information in memory
// is not a cache to try to optimize the query to the original repository but try caching the
// information of the rooms that are asked frequently and save most of the room info accesses.
func NewCachedUserRepository(next UserRepository) (UserRepository, error) {
	cui, err := lru.New[string, *model.User](500)
	if err != nil {
		return nil, fmt.Errorf("could not initialize cache")
	}
	cun, err := lru.New[string, bool](500)
	if err != nil {
		return nil, fmt.Errorf("could not initialize cache")
	}

	cu, err := lru.New[string, *model.User](500)
	if err != nil {
		return nil, fmt.Errorf("could not initialize cache")
	}

	return &cachedUserRepository{
		userIDCache:         cui,
		userNameExistsCache: cun,
		userNameCache:       cu,
		UserRepository:      next,
	}, nil
}

func (c cachedUserRepository) GetUserByID(ctx context.Context, userID string) (u *model.User, err error) {
	user, ok := c.userIDCache.Get(userID)
	if ok {
		return user, nil
	}

	user, err = c.UserRepository.GetUserByID(ctx, userID)
	if err != nil {
		return user, err
	}

	// Save in cache.
	_ = c.userIDCache.Add(userID, user)

	return user, err
}

func (c cachedUserRepository) UserExists(ctx context.Context, userID string) (ex bool, err error) {
	// Try a best effort.
	_, ok := c.userIDCache.Get(userID)
	if ok {
		return true, nil
	}

	return c.UserRepository.UserExists(ctx, userID)
}

func (c cachedUserRepository) UserExistsByNameInsensitive(ctx context.Context, roomID, username string) (ex bool, err error) {
	username = strings.ToLower(username)
	k := roomID + username

	_, ok := c.userNameExistsCache.Get(k)
	if ok {
		return true, nil
	}

	ex, err = c.UserRepository.UserExistsByNameInsensitive(ctx, roomID, username)
	if !ex || err != nil {
		return ex, err
	}

	// Save in cache.
	_ = c.userNameExistsCache.Add(k, true)

	return ex, err
}

// GetUserByNameInsensitive returns the user using user ID being insensitive.
func (c cachedUserRepository) GetUserByNameInsensitive(ctx context.Context, roomID, username string) (*model.User, error) {
	username = strings.ToLower(username)
	k := roomID + username

	u, ok := c.userNameCache.Get(k)
	if ok {
		return u, nil
	}

	us, err := c.UserRepository.GetUserByNameInsensitive(ctx, roomID, username)
	if err != nil {
		return us, err
	}

	// Save in cache.
	_ = c.userNameCache.Add(k, us)

	return us, err
}
