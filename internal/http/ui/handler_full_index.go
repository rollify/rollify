package ui

import (
	"net/http"
)

func (u ui) handlerFullIndex() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.tplRenderer.RenderResponse(r.Context(), w, "index", tplDataCreateRoom{})
	})
}
