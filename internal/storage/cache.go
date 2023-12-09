package storage

import (
	"context"
	"fmt"

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
