package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/r3labs/sse/v2"
	gohttmetrics "github.com/slok/go-http-metrics/middleware"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
)

var (
	//go:embed all:static
	staticFS embed.FS
	//go:embed all:templates
	templatesFS embed.FS
)

// Config is the configuration to serve the API.
type Config struct {
	DiceAppService  dice.Service
	RoomAppService  room.Service
	UserAppService  user.Service
	MetricsRecorder MetricsRecorder
	ServerPrefix    string
	TimeNow         func() time.Time
	SSEServer       *sse.Server
	Logger          log.Logger
}

func (c *Config) defaults() error {
	if c.DiceAppService == nil {
		return fmt.Errorf("dice.Service application service is required")
	}

	if c.RoomAppService == nil {
		return fmt.Errorf("room.Service application service is required")
	}

	if c.UserAppService == nil {
		return fmt.Errorf("user.Service application service is required")
	}

	if c.SSEServer == nil {
		return fmt.Errorf("an SSE server is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{
		"http": "ui",
	})

	if c.MetricsRecorder == nil {
		c.MetricsRecorder = noopMetricsRecorder
		c.Logger.Warningf("metrics recorder disabled")
	}

	if c.ServerPrefix == "" {
		c.ServerPrefix = "/u"
	}

	if c.TimeNow == nil {
		c.TimeNow = time.Now
	}

	return nil
}

type ui struct {
	diceAppSvc        dice.Service
	roomAppSvc        room.Service
	userAppSvc        user.Service
	router            chi.Router
	servePrefix       string
	logger            log.Logger
	templates         *template.Template
	staticFS          fs.FS
	metricsMiddleware gohttmetrics.Middleware
	sseServer         *sse.Server
	timeNow           func() time.Time
}

// New returns UI handler.
func New(cfg Config) (http.Handler, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("wrong configuration: %w", err)
	}

	templates, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse templates: %w", err)
	}

	sanitizedStaticFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("could not sanitize static FS: %w", err)
	}

	a := ui{
		diceAppSvc:  cfg.DiceAppService,
		roomAppSvc:  cfg.RoomAppService,
		userAppSvc:  cfg.UserAppService,
		router:      chi.NewRouter(),
		servePrefix: cfg.ServerPrefix,
		templates:   templates,
		staticFS:    sanitizedStaticFS,
		logger:      cfg.Logger,
		metricsMiddleware: gohttmetrics.New(gohttmetrics.Config{
			Recorder: cfg.MetricsRecorder,
			Service:  "ui",
		}),
		sseServer: cfg.SSEServer,
		timeNow:   cfg.TimeNow,
	}

	a.registerRoutes()

	return a, nil
}

func (u ui) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	router.Mount(u.servePrefix, u.router)
	router.ServeHTTP(w, r)
}
