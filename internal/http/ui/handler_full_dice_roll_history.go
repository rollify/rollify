package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
)

const maxDiceResults = 10

type userDiceRoll struct {
	Username     string
	TS           string
	DiceResults  []diceResult
	NextItemsURL string
	IsPushUpdate bool
}

type tplDatadiceRollHistory struct {
	tplDataCommon
	RoomName       string
	RoomID         string
	NewDiceRollURL string
	IsDiceHistory  bool
	SSEURL         string
	Dice           []die
	Results        []userDiceRoll
}

func (u ui) handlerFullDiceRollHistory() http.HandlerFunc {
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

		res, err := u.diceAppSvc.ListDiceRolls(r.Context(), dice.ListDiceRollsRequest{
			UserID:   userID,
			RoomID:   roomID,
			PageOpts: model.PaginationOpts{Size: maxDiceResults},
		})
		if err != nil {
			u.handleError(w, fmt.Errorf("could list dice rolls: %w", err))
			return
		}

		u.tplRender(w, "room_dice_roll_history", tplDatadiceRollHistory{
			tplDataCommon: tplDataCommon{
				URLPrefix: u.servePrefix,
			},
			RoomName:       room.Room.Name,
			RoomID:         room.Room.Name,
			NewDiceRollURL: u.servePrefix + "/room/" + room.Room.ID,
			IsDiceHistory:  true,
			Dice:           []die{dieD4, dieD6, dieD8, dieD10, dieD12, dieD20},
			Results:        u.formatDiceHistory(*res, roomID),
			SSEURL:         fmt.Sprintf("%s/subscribe/room/dice-roll-history?%s=%s%s", u.servePrefix, queryParamSSEStream, sseStreamPrefixHTML, roomID),
		})
	})
}

func (u ui) formatDiceHistory(m dice.ListDiceRollsResponse, roomID string) []userDiceRoll {
	res := []userDiceRoll{}
	for _, d := range m.DiceRolls {
		res = append(res, u.mapDiceRollToTplModel(d, false))
	}

	// On last one add pagination.
	if m.Cursors.HasNext {
		url := fmt.Sprintf("%s/room/%s/dice-roll-history/more-items?%s=%s", u.servePrefix, roomID, queryParamCursor, m.Cursors.LastCursor)
		res[len(res)-1].NextItemsURL = url
	}

	return res
}

func (u ui) mapDiceRollToTplModel(d model.DiceRoll, isPush bool) userDiceRoll {
	groupedResults := map[string][]uint{}
	for _, r := range d.Dice {
		groupedResults[r.Type.ID()] = append(groupedResults[r.Type.ID()], r.Side)
	}

	// TODO(slok): Sort.
	return userDiceRoll{
		Username: d.UserID,
		TS:       fmt.Sprintf("%v", time.Since(d.CreatedAt).Round(time.Second)),
		DiceResults: []diceResult{
			{Dice: dieD4, Results: groupedResults[dieD4.ID()]},
			{Dice: dieD6, Results: groupedResults[dieD6.ID()]},
			{Dice: dieD8, Results: groupedResults[dieD8.ID()]},
			{Dice: dieD10, Results: groupedResults[dieD10.ID()]},
			{Dice: dieD12, Results: groupedResults[dieD12.ID()]},
			{Dice: dieD20, Results: groupedResults[dieD20.ID()]},
		},
		IsPushUpdate: isPush,
	}
}
