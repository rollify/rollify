package ui

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
)

type diceResult struct {
	Dice    die
	Results []uint
}

type tplDataNewDiceRoll struct {
	tplDataCommon
	DiceResult []diceResult
}

func (u ui) handlerSnippetNewDiceRoll() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, urlParamRoomID)
		userID := cookies.GetUserID(r, roomID)

		// Get result from dice.
		err := r.ParseForm()
		if err != nil {
			u.handleError(w, fmt.Errorf("could create dice roll: %w", err))
			return
		}

		ds := []model.DieType{}
		if q, err := strconv.Atoi(r.FormValue(dieD4.ID())); err == nil {
			ds = addDice(ds, dieD4.DieType, q)
		}
		if q, err := strconv.Atoi(r.FormValue(dieD6.ID())); err == nil {
			ds = addDice(ds, dieD6.DieType, q)
		}
		if q, err := strconv.Atoi(r.FormValue(dieD8.ID())); err == nil {
			ds = addDice(ds, dieD8.DieType, q)
		}
		if q, err := strconv.Atoi(r.FormValue(dieD10.ID())); err == nil {
			ds = addDice(ds, dieD10.DieType, q)
		}
		if q, err := strconv.Atoi(r.FormValue(dieD12.ID())); err == nil {
			ds = addDice(ds, dieD12.DieType, q)
		}
		if q, err := strconv.Atoi(r.FormValue(dieD20.ID())); err == nil {
			ds = addDice(ds, dieD20.DieType, q)
		}
		res, err := u.diceAppSvc.CreateDiceRoll(r.Context(), dice.CreateDiceRollRequest{
			UserID: userID,
			RoomID: roomID,
			Dice:   ds,
		})
		if err != nil {
			u.handleError(w, fmt.Errorf("could create dice roll: %w", err))
			return
		}

		// Bake result.
		groupedResults := map[string][]uint{}
		for _, res := range res.DiceRoll.Dice {
			groupedResults[res.Type.ID()] = append(groupedResults[res.Type.ID()], res.Side)
		}

		for _, v := range groupedResults {
			slices.Sort(v)
		}

		drs := []diceResult{
			{Dice: dieD4, Results: groupedResults[dieD4.ID()]},
			{Dice: dieD6, Results: groupedResults[dieD6.ID()]},
			{Dice: dieD8, Results: groupedResults[dieD8.ID()]},
			{Dice: dieD10, Results: groupedResults[dieD10.ID()]},
			{Dice: dieD12, Results: groupedResults[dieD12.ID()]},
			{Dice: dieD20, Results: groupedResults[dieD20.ID()]},
		}

		d := tplDataNewDiceRoll{
			tplDataCommon: tplDataCommon{URLPrefix: u.servePrefix},
			DiceResult:    drs,
		}
		u.tplRender(w, "dice_roll_result", d)

	})
}

func addDice(ds []model.DieType, t model.DieType, q int) []model.DieType {
	for i := 0; i < q; i++ {
		ds = append(ds, t)
	}
	return ds
}
