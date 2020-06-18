package apiv1

import (
	"fmt"
	"time"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
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
	ID string `json:"id"`
	// Representation in RFC3339.
	CreateAt string    `json:"created_at"`
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Dice     []dieRoll `json:"dice"`
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
		ID:       r.DiceRoll.ID,
		CreateAt: r.DiceRoll.CreatedAt.Format(time.RFC3339),
		RoomID:   r.DiceRoll.RoomID,
		UserID:   r.DiceRoll.UserID,
		Dice:     ds,
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

type listDiceRollsResponse struct {
	Items []diceRollResponse `json:"items"`
}

type diceRollResponse struct {
	ID string `json:"id"`
	// Representation in RFC3339.
	CreateAt string            `json:"created_at"`
	UserID   string            `json:"user_id"`
	RoomID   string            `json:"room_id"`
	Dice     []dieRollResponse `json:"dice"`
}

type dieRollResponse struct {
	ID     string `json:"id"`
	TypeID string `json:"type_id"`
	Side   uint   `json:"side"`
}

type listDiceRollsRequest struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
}

func mapModelToAPIListDiceRolls(r dice.ListDiceRollsResponse) listDiceRollsResponse {
	items := make([]diceRollResponse, 0, len(r.DiceRolls))
	for _, dr := range r.DiceRolls {
		ds := make([]dieRollResponse, 0, len(dr.Dice))
		for _, d := range dr.Dice {
			ds = append(ds, dieRollResponse{
				ID:     d.ID,
				TypeID: d.Type.ID(),
				Side:   d.Side,
			})
		}
		items = append(items, diceRollResponse{
			ID:       dr.ID,
			CreateAt: dr.CreatedAt.Format(time.RFC3339),
			RoomID:   dr.RoomID,
			UserID:   dr.UserID,
			Dice:     ds,
		})
	}

	return listDiceRollsResponse{
		Items: items,
	}
}

func mapAPIToModelListDiceRolls(r listDiceRollsRequest) (*dice.ListDiceRollsRequest, error) {
	if r.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	return &dice.ListDiceRollsRequest{
		UserID: r.UserID,
		RoomID: r.RoomID,
	}, nil
}

type createRoomResponse struct {
	ID string `json:"id"`
	// Representation in RFC3339.
	CreateAt string `json:"created_at"`
	Name     string `json:"name"`
}
type createRoomRequest struct {
	Name string `json:"name"`
}

func mapModelToAPICreateRoom(r room.CreateRoomResponse) createRoomResponse {
	return createRoomResponse{
		ID:       r.Room.ID,
		CreateAt: r.Room.CreatedAt.Format(time.RFC3339),
		Name:     r.Room.Name,
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

type createUserResponse struct {
	ID string `json:"id"`
	// Representation in RFC3339.
	CreateAt string `json:"created_at"`
	Name     string `json:"name"`
	RoomID   string `json:"room_id"`
}
type createUserRequest struct {
	Name   string `json:"name"`
	RoomID string `json:"room_id"`
}

func mapModelToAPICreateUser(r user.CreateUserResponse) createUserResponse {
	return createUserResponse{
		ID:       r.User.ID,
		CreateAt: r.User.CreatedAt.Format(time.RFC3339),
		Name:     r.User.Name,
		RoomID:   r.User.RoomID,
	}
}

func mapAPIToModelCreateUser(r createUserRequest) (*user.CreateUserRequest, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if r.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	return &user.CreateUserRequest{
		Name:   r.Name,
		RoomID: r.RoomID,
	}, nil
}

type listUsersResponse struct {
	Items []userResponse `json:"items"`
}

type userResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	RoomID string `json:"room_id"`
	// Representation in RFC3339.
	CreateAt string `json:"created_at"`
}

type listUsersRequest struct {
	RoomID string `json:"room_id"`
}

func mapModelToAPIListUsers(r user.ListUsersResponse) listUsersResponse {
	items := make([]userResponse, 0, len(r.Users))
	for _, u := range r.Users {
		items = append(items, userResponse{
			ID:       u.ID,
			Name:     u.Name,
			RoomID:   u.RoomID,
			CreateAt: u.CreatedAt.Format(time.RFC3339),
		})
	}
	return listUsersResponse{
		Items: items,
	}
}

func mapAPIToModelListUsers(r listUsersRequest) (*user.ListUsersRequest, error) {
	if r.RoomID == "" {
		return nil, fmt.Errorf("room_id is required")
	}

	return &user.ListUsersRequest{
		RoomID: r.RoomID,
	}, nil
}
