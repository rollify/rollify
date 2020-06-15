package model

import "time"

// DieRoll represents a single die.
type DieRoll struct {
	// ID is the ID of the Die roll
	ID string
	// Type is the Die type, e.g: d6, d20.
	Type DieType
	// Side is the side we got after a die roll.
	Side uint
}

// DiceRoll represents a dice roll.
type DiceRoll struct {
	// ID is the ID of the DiceRoll.
	ID string
	// Serial is an incremental ID (managed by the storage) that is safe to sort.
	// Is encouraged to use this to sort or know a point in time (after/before) instead
	// of CreatedAt.
	Serial uint
	// Created at is when the dice roll was created.
	CreatedAt time.Time
	// UserID is the ID of the room were the dice roll was made.
	RoomID string
	// UserID is the ID of the user that made the dice roll.
	UserID string
	// Dice are the rolled dice values involved in the dice roll.
	Dice []DieRoll
}
