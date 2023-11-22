package ui

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rollify/rollify/internal/http/ui/htmx"
	"github.com/rollify/rollify/internal/room"
)

const (
	formFieldCreateRoomRoomName = "roomName"
)

type tplDataCreateRoom struct {
	tplDataCommon
	FormErrors []string
}

func (u ui) createRoom() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form.
		roomName := r.FormValue(formFieldCreateRoomRoomName)
		roomName = strings.TrimSpace(roomName)
		if roomName == "" {
			d := tplDataCreateRoom{
				tplDataCommon: u.tplCommonData(),
				FormErrors:    []string{"Room name can't be empty"},
			}
			u.tplRender(w, "create_room_form", d)
			return
		}

		// Create the room.
		resp, err := u.roomAppSvc.CreateRoom(r.Context(), room.CreateRoomRequest{Name: roomName})
		if err != nil {
			u.handleError(w, fmt.Errorf("could not create room: %w", err))
			return
		}

		// Room created, redirect to the room login.
		htmx.NewResponse().WithRedirect(u.servePrefix + "/login-room/" + resp.Room.ID).SetHeaders(w)
	})
}
