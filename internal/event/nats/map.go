package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rollify/rollify/internal/model"
)

type eventDiceRollCreated struct {
	DiceRoll diceRoll
}

type diceRoll struct {
	ID        string
	Serial    uint
	CreatedAt time.Time
	RoomID    string
	UserID    string
	Dice      []dieRoll
}

type dieRoll struct {
	ID   string
	Type string
	Side uint
}

// maps model to a stream model used to be shared.
func mapModelToBytesEventDiceRollCreated(e model.EventDiceRollCreated) ([]byte, error) {
	res := eventDiceRollCreated{
		DiceRoll: diceRoll{
			ID:        e.DiceRoll.ID,
			Serial:    e.DiceRoll.Serial,
			CreatedAt: e.DiceRoll.CreatedAt,
			RoomID:    e.DiceRoll.RoomID,
			UserID:    e.DiceRoll.UserID,
			Dice:      make([]dieRoll, 0, len(e.DiceRoll.Dice)),
		},
	}

	for _, dr := range e.DiceRoll.Dice {
		res.DiceRoll.Dice = append(res.DiceRoll.Dice, dieRoll{
			ID:   dr.ID,
			Type: dr.Type.ID(),
			Side: dr.Side,
		})
	}

	bs, err := json.Marshal(&res)
	if err != nil {
		return nil, fmt.Errorf("could not marshall event to bytes: %w", err)
	}

	return bs, nil
}

func mapBytesToModelEventDiceRollCreated(data []byte) (*model.EventDiceRollCreated, error) {
	e := &eventDiceRollCreated{}
	err := json.Unmarshal(data, e)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall bytes to event: %w", err)
	}

	res := &model.EventDiceRollCreated{
		DiceRoll: model.DiceRoll{
			ID:        e.DiceRoll.ID,
			Serial:    e.DiceRoll.Serial,
			CreatedAt: e.DiceRoll.CreatedAt,
			RoomID:    e.DiceRoll.RoomID,
			UserID:    e.DiceRoll.UserID,
			Dice:      make([]model.DieRoll, 0, len(e.DiceRoll.Dice)),
		},
	}

	for _, dr := range e.DiceRoll.Dice {
		dt, ok := model.DiceTypes[dr.Type]
		if !ok {
			return nil, fmt.Errorf("%s die type is not valid", dr.Type)
		}
		res.DiceRoll.Dice = append(res.DiceRoll.Dice, model.DieRoll{
			ID:   dr.ID,
			Type: dt,
			Side: dr.Side,
		})
	}

	return res, nil
}
