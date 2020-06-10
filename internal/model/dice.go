package model

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
	// Dice are the rolled dice values involved in the dice roll.
	Dice []DieRoll
}
