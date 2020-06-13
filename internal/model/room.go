package model

import "time"

// Room represents a room.
type Room struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
