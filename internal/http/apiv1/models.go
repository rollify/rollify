package apiv1

import "github.com/rollify/rollify/internal/dice"

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
