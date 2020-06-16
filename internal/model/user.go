package model

import "time"

// User represents a user. Users are unique by room, but a same phisical
// person can be in different rooms with different users.
type User struct {
	ID        string
	Name      string
	RoomID    string
	CreatedAt time.Time
}
