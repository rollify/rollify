package ui

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"maps"
	"net/http"
	"text/template"

	"github.com/rollify/rollify/internal/log"
)

var (
	//go:embed all:static
	staticFS embed.FS
	//go:embed all:templates
	templatesFS embed.FS
)

const (
	commonDataKeyRoomID    = "RoomID"
	commonDataKeyURLPrefix = "URLPrefix"
	commonDataKeyErrors    = "Errors"
)

// tplRenderer is a util that will make rendering templates easier and standarize inside the server.
type tplRenderer struct {
	logger log.Logger
	tpls   *template.Template

	// Extra data.
	// This data will be available on all templates as `Common.{KEY}`.
	CommonData map[string]any
}

func NewTplRenderer(logger log.Logger) (*tplRenderer, error) {
	templates, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse templates: %w", err)
	}

	return &tplRenderer{
		logger: logger,
		tpls:   templates,
		CommonData: map[string]any{
			commonDataKeyRoomID:    "",
			commonDataKeyURLPrefix: "",
			commonDataKeyErrors:    []string{},
		},
	}, nil
}

func (t *tplRenderer) WithURLPrefix(serverPrefix string) *tplRenderer {
	c := maps.Clone(t.CommonData)
	c[commonDataKeyURLPrefix] = serverPrefix

	return &tplRenderer{
		logger:     t.logger,
		tpls:       t.tpls,
		CommonData: c,
	}
}

func (t *tplRenderer) withRoom(roomID string) *tplRenderer {
	c := maps.Clone(t.CommonData)
	c[commonDataKeyRoomID] = roomID

	return &tplRenderer{
		logger:     t.logger,
		tpls:       t.tpls,
		CommonData: c,
	}
}

func (t *tplRenderer) WithErrors(errors []string) *tplRenderer {
	c := maps.Clone(t.CommonData)
	c[commonDataKeyErrors] = errors

	return &tplRenderer{
		logger:     t.logger,
		tpls:       t.tpls,
		CommonData: c,
	}
}

func (t *tplRenderer) RenderResponse(ctx context.Context, w http.ResponseWriter, tplName string, data any) {
	d := struct {
		Common map[string]any
		Data   any
	}{
		Common: t.CommonData,
		Data:   data,
	}
	err := t.tpls.ExecuteTemplate(w, tplName, d)
	if err != nil {
		t.logger.Errorf("Could not render template: %s", err)
		// TODO(slok): Render 500 template.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *tplRenderer) Render(ctx context.Context, tplName string, data any) (string, error) {
	d := struct {
		Common map[string]any
		Data   any
	}{
		Common: t.CommonData,
		Data:   data,
	}
	var b bytes.Buffer
	err := t.tpls.ExecuteTemplate(&b, tplName, d)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
