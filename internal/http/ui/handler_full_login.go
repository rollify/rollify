package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
)

func (u ui) handlerFullLogin() http.HandlerFunc {
	type tplData struct {
		Users    []model.User
		RoomName string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, urlParamRoomID)

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

		u.tplRenderer.withRoom(roomID).RenderResponse(r.Context(), w, "login", tplData{
			Users:    mresp.Users,
			RoomName: room.Room.Name,
		})
	})
}
