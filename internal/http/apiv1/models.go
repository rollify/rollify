package apiv1

import (
	"fmt"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
)

type listDiceTypesResponse struct {
	Items []diceTypeResponse `json:"items"`
}

type diceTypeResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Sides int    `json:"sides"`
}

func mapModelToAPIListDiceTypes(r dice.ListDiceTypesResponse) listDiceTypesResponse {
	dt := make([]diceTypeResponse, 0, len(r.DiceTypes))
	for _, d := range r.DiceTypes {
		dt = append(dt, diceTypeResponse{
			ID:    d.ID(),
			Name:  d.ID(),
			Sides: int(d.Sides()),
		})
	}
	return listDiceTypesResponse{
		Items: dt,
	}
}

type createDiceRollResponse struct {
	ID   string    `json:"id"`
	Dice []dieRoll `json:"dice"`
}

type dieRoll struct {
	ID         string `json:"id"`
	DiceTypeID string `json:"dice_type_id"`
	Side       uint   `json:"side"`
}

type createDiceRollRequest struct {
	UserID      string   `json:"user_id"`
	RoomID      string   `json:"room_id"`
	DiceTypeIDs []string `json:"dice_type_ids"`
}

func mapModelToAPIcreateDiceRoll(r dice.CreateDiceRollResponse) createDiceRollResponse {
	ds := make([]dieRoll, 0, len(r.DiceRoll.Dice))
	for _, d := range r.DiceRoll.Dice {
		ds = append(ds, dieRoll{
			ID:         d.ID,
			DiceTypeID: d.Type.ID(),
			Side:       d.Side,
		})
	}
	return createDiceRollResponse{
		ID:   r.DiceRoll.ID,
		Dice: ds,
	}
}

func mapAPIToModelcreateDiceRoll(r createDiceRollRequest) (*dice.CreateDiceRollRequest, error) {
	if r.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if r.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	if len(r.DiceTypeIDs) == 0 {
		return nil, fmt.Errorf("dice_type_ids are required")
	}

	dts := make([]model.DieType, 0, len(r.DiceTypeIDs))
	for _, id := range r.DiceTypeIDs {
		dt, ok := model.DiceTypes[id]
		if !ok {
			return nil, fmt.Errorf("%s die type is not valid", id)
		}
		dts = append(dts, dt)
	}

	return &dice.CreateDiceRollRequest{
		UserID: r.UserID,
		RoomID: r.RoomID,
		Dice:   dts,
	}, nil
}

type createRoomResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type createRoomRequest struct {
	Name string `json:"name"`
}

func mapModelToAPIcreateRoom(r room.CreateRoomResponse) createRoomResponse {
	return createRoomResponse{
		ID:   r.Room.ID,
		Name: r.Room.Name,
	}
}

func mapAPIToModelCreateRoom(r createRoomRequest) (*room.CreateRoomRequest, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return &room.CreateRoomRequest{
		Name: r.Name,
	}, nil
}
