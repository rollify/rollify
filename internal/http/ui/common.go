package ui

import (
	"net/http"

	"github.com/rollify/rollify/internal/http/ui/htmx"
)

func (u ui) handleError(w http.ResponseWriter, err error) {
	u.logger.Errorf("HTTP handler error: %s", err)
	w.WriteHeader(http.StatusInternalServerError)
}

func (u ui) redirectToURL(w http.ResponseWriter, r *http.Request, url string) {
	// If HTMX request, redirect with HTMX, if not regular redirect.
	if htmx.NewRequest(r.Header).IsHTMXRequest() {
		htmx.NewResponse().WithRedirect(url).SetHeaders(w)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
