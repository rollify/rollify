package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
)

type tplDataLoginRoom struct {
	tplDataCommon
	Users    []model.User
	RoomID   string
	RoomName string
}

func (u ui) loginRoom() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, paramRoomID)

		// Get room name.
		room, err := u.roomAppSvc.GetRoom(r.Context(), room.GetRoomRequest{ID: roomID})
		if err != nil {
			u.handleError(w, fmt.Errorf("could not get room: %w", err))
			return
		}

		// List the available users.
		mresp, err := u.userAppSvc.ListUsers(r.Context(), user.ListUsersRequest{RoomID: roomID})
		if err != nil {
			u.handleError(w, fmt.Errorf("could not list users: %w", err))
			return
		}

		d := tplDataLoginRoom{
			tplDataCommon: u.tplCommonData(),
			Users:         mresp.Users,
			RoomID:        roomID,
			RoomName:      room.Room.Name,
		}
		u.tplRender(w, "login_room", d)
	})
}
