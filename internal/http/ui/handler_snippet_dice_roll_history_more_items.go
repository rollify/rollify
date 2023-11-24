package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
)

type tplDatadiceRollHistoryMoreItems struct {
	Results []userDiceRoll
}

func (u ui) handlerSnippetDiceRollHistoryMoreItems() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get(queryParamCursor)
		roomID := chi.URLParam(r, urlParamRoomID)
		userID := cookies.GetUserID(r, roomID)

		// If not user ID, redirect to room selection.
		if userID == "" {
			u.redirectToURL(w, r, u.servePrefix+"/login/"+roomID)
			return
		}

		res, err := u.diceAppSvc.ListDiceRolls(r.Context(), dice.ListDiceRollsRequest{
			UserID:   userID,
			RoomID:   roomID,
			PageOpts: model.PaginationOpts{Cursor: cursor, Size: maxDiceResults},
		})
		if err != nil {
			u.handleError(w, fmt.Errorf("could list dice rolls: %w", err))
			return
		}

		u.tplRender(w, "dice_roll_history_rows", tplDatadiceRollHistoryMoreItems{
			Results: u.formatDiceHistory(*res, roomID),
		})
	})
}
