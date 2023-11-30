package ui

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
)

const maxDiceResults = 10

type userDiceRoll struct {
	Username     string
	UnixTS       int64
	PrettyTS     string
	DiceResults  []diceResult
	IsPushUpdate bool
}

func (u ui) handlerFullDiceRollHistory() http.HandlerFunc {
	type tplData struct {
		RoomName       string
		RoomID         string
		NewDiceRollURL string
		IsDiceHistory  bool
		SSEURL         string
		Dice           []die
		Results        []userDiceRoll
		NextItemsURL   string
	}

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
			RoomID:   roomID,
			PageOpts: model.PaginationOpts{Size: maxDiceResults},
		})
		if err != nil {
			u.handleError(w, fmt.Errorf("could list dice rolls: %w", err))
			return
		}

		roomUsers, err := u.userAppSvc.ListUsers(r.Context(), user.ListUsersRequest{RoomID: roomID})
		if err != nil {
			u.handleError(w, fmt.Errorf("could list room users: %w", err))
			return
		}

		nextItemsURL := ""
		if res.Cursors.HasNext {
			nextItemsURL = fmt.Sprintf("%s/room/%s/dice-roll-history/more-items?%s=%s", u.servePrefix, roomID, queryParamCursor, res.Cursors.LastCursor)
		}

		u.tplRenderer.withRoom(roomID).RenderResponse(r.Context(), w, "room_dice_roll_history", tplData{
			RoomName:       room.Room.Name,
			RoomID:         room.Room.Name,
			NewDiceRollURL: u.servePrefix + "/room/" + room.Room.ID,
			IsDiceHistory:  true,
			Dice:           []die{dieD4, dieD6, dieD8, dieD10, dieD12, dieD20},
			Results:        u.formatDiceHistory(*res, roomUsers.Users),
			SSEURL:         fmt.Sprintf("%s/subscribe/room/dice-roll-history?%s=%s%s", u.servePrefix, queryParamSSEStream, sseStreamPrefixHTML, roomID),
			NextItemsURL:   nextItemsURL,
		})
	})
}

func (u ui) formatDiceHistory(m dice.ListDiceRollsResponse, users []model.User) []userDiceRoll {
	us := map[string]model.User{}
	for _, u := range users {
		us[u.ID] = u
	}
	res := []userDiceRoll{}
	for _, d := range m.DiceRolls {
		res = append(res, u.mapDiceRollToTplModel(d, us[d.UserID], false))
	}

	return res
}

func (u ui) mapDiceRollToTplModel(d model.DiceRoll, user model.User, isPush bool) userDiceRoll {
	groupedResults := map[string][]uint{}
	for _, r := range d.Dice {
		groupedResults[r.Type.ID()] = append(groupedResults[r.Type.ID()], r.Side)
	}

	for _, v := range groupedResults {
		slices.Sort(v)
	}

	return userDiceRoll{
		Username: user.Name,
		UnixTS:   d.CreatedAt.UTC().Unix(),
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
