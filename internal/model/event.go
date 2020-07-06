package model

// Event represents an app event.
type Event interface {
	Type() string
}

// EventDiceRollCreated is a dice roll creation event.
type EventDiceRollCreated struct {
	DiceRoll DiceRoll
}

// Type satisfies Event interface.
func (EventDiceRollCreated) Type() string { return "EventDiceRollCreated" }
