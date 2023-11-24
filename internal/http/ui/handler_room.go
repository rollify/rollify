package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rollify/rollify/internal/room"
)

type tplDataRoom struct {
	tplDataCommon
	RoomName       string
	RoomID         string
	Dice           []die
	DiceHistoryURL string
	IsDiceHistory  bool
	SSEURL         string
}

func (u ui) room() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, urlParamRoomID)
		userID := cookies.GetUserID(r, roomID)

		// If not user ID, redirect to room selection.
		if userID == "" {
			u.redirectToURL(w, r, u.servePrefix+"/login/"+roomID)
			return
		}

		room, err := u.roomAppSvc.GetRoom(r.Context(), room.GetRoomRequest{ID: roomID})
		if err != nil {
			u.handleError(w, fmt.Errorf("could not get room: %w", err))
			return
		}

		u.tplRender(w, "room", tplDataRoom{
			tplDataCommon: tplDataCommon{
				URLPrefix: u.servePrefix,
			},
			RoomName:       room.Room.Name,
			RoomID:         room.Room.ID,
			DiceHistoryURL: u.servePrefix + "/room/" + room.Room.ID + "/dice-roll-history",
			Dice: []die{
				dieD4,
				dieD6,
				dieD8,
				dieD10,
				dieD12,
				dieD20,
			},
			IsDiceHistory: false,
			SSEURL:        fmt.Sprintf("%s/subscribe/room/dice-roll-history?%s=%s%s", u.servePrefix, queryParamSSEStream, sseStreamPrefixNotification, roomID),
		})
	})
}
