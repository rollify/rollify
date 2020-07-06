package apiv1

import (
	"fmt"
	"net/url"
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
	Meta  metadata           `json:"metadata"`
}

type metadata struct {
	FirstCursor string `json:"first_cursor"`
	LastCursor  string `json:"last_cursor"`
	HasNext     bool   `json:"has_next"`
	HasPrevious bool   `json:"has_previous"`
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
		Meta: metadata{
			FirstCursor: r.Cursors.FirstCursor,
			LastCursor:  r.Cursors.LastCursor,
			HasNext:     r.Cursors.HasNext,
			HasPrevious: r.Cursors.HasPrevious,
		},
	}
}

const (
	listDiceRollsParamUserID      = "user-id"
	listDiceRollsParamRoomID      = "room-id"
	listDiceRollsPaginationCursor = "cursor"
	listDiceRollsPaginationOrder  = "order"
)

func mapAPIToModelPaginationOrder(order string) (model.PaginationOrder, error) {
	switch order {
	case "":
		return model.PaginationOrderDefault, nil
	case "asc":
		return model.PaginationOrderAsc, nil
	case "desc":
		return model.PaginationOrderDesc, nil
	default:
		return 0, fmt.Errorf("pagination order '%s' is invalid", order)
	}
}

func mapAPIToModelListDiceRolls(p url.Values) (*dice.ListDiceRollsRequest, error) {
	roomID := p.Get(listDiceRollsParamRoomID)
	if roomID == "" {
		return nil, fmt.Errorf("room-id is required")
	}

	userID := p.Get(listDiceRollsParamUserID)
	cursor := p.Get(listDiceRollsPaginationCursor)
	order := p.Get(listDiceRollsPaginationOrder)
	mOrder, err := mapAPIToModelPaginationOrder(order)
	if err != nil {
		return nil, err
	}

	return &dice.ListDiceRollsRequest{
		UserID: userID,
		RoomID: roomID,
		PageOpts: model.PaginationOpts{
			Cursor: cursor,
			Order:  mOrder,
		},
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

type getRoomResponse struct {
	ID string `json:"id"`
	// Representation in RFC3339.
	CreateAt string `json:"created_at"`
	Name     string `json:"name"`
}

func mapModelToAPIGetRoom(r room.GetRoomResponse) getRoomResponse {
	return getRoomResponse{
		ID:       r.Room.ID,
		CreateAt: r.Room.CreatedAt.Format(time.RFC3339),
		Name:     r.Room.Name,
	}
}

const getRoomParamRoomID = "id"

func mapAPIToModelGetRoom(params map[string]string) (*room.GetRoomRequest, error) {
	id, ok := params[getRoomParamRoomID]
	if !ok {
		return nil, fmt.Errorf("room id is required")
	}

	return &room.GetRoomRequest{
		ID: id,
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
	ID   string `json:"id"`
	Name string `json:"name"`
	// Representation in RFC3339.
	CreateAt string `json:"created_at"`
}

func mapModelToAPIListUsers(r user.ListUsersResponse) listUsersResponse {
	items := make([]userResponse, 0, len(r.Users))
	for _, u := range r.Users {
		items = append(items, userResponse{
			ID:       u.ID,
			Name:     u.Name,
			CreateAt: u.CreatedAt.Format(time.RFC3339),
		})
	}
	return listUsersResponse{
		Items: items,
	}
}

const listUsersParamRoomID = "room-id"

func mapAPIToModelListUsers(p url.Values) (*user.ListUsersRequest, error) {
	roomID := p.Get(listUsersParamRoomID)
	if roomID == "" {
		return nil, fmt.Errorf("room-id is required")
	}

	return &user.ListUsersRequest{
		RoomID: roomID,
	}, nil
}

type wsEventMeta struct {
	Type string `json:"type"`
}

type wsDiceRollCreatedEvent struct {
	Metadata wsEventMeta `json:"metadata"`
}

func mapModelToAPIWSDiceRollCreatedEvent(e model.EventDiceRollCreated) wsDiceRollCreatedEvent {
	return wsDiceRollCreatedEvent{
		Metadata: wsEventMeta{
			Type: "diceRollCreated",
		},
	}
}
