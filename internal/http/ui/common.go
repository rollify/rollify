package ui

import (
	"bytes"
	"net/http"

	"github.com/rollify/rollify/internal/http/ui/htmx"
)

func (u ui) tplRender(w http.ResponseWriter, templateName string, data any) {
	err := u.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u ui) tplRenderS(templateName string, data any) string {
	var b bytes.Buffer
	err := u.templates.ExecuteTemplate(&b, templateName, data)
	if err != nil {
		u.logger.Errorf("Error rendering template: %s", err)
	}

	return b.String()
}

// tplDataCommon is a common data context that can be used on all templates and
// any template can use it.
type tplDataCommon struct {
	URLPrefix      string
	Errors         []string
	DiceHistoryURL string
}

func (u ui) tplCommonData() tplDataCommon {
	return tplDataCommon{
		URLPrefix: u.servePrefix,
		Errors:    []string{},
	}
}

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
