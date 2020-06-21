package apiv1

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/user"
)

// Config is the configuration to serve the API.
type Config struct {
	DiceAppService dice.Service
	RoomAppService room.Service
	UserAppService user.Service
	ServePefix     string
	Logger         log.Logger
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

	if c.ServePefix == "" {
		c.ServePefix = "/api/v1"
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{
		"http":   "apiv1",
		"prefix": c.ServePefix,
	})

	return nil
}

type apiv1 struct {
	diceAppSvc    dice.Service
	roomAppSvc    room.Service
	UserAppSvc    user.Service
	logger        log.Logger
	apiws         *restful.WebService
	restContainer *restful.Container
}

// New returns API v1 HTTP handler.
func New(cfg Config) (http.Handler, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("wrong configuration: %w", err)
	}

	a := apiv1{
		diceAppSvc: cfg.DiceAppService,
		roomAppSvc: cfg.RoomAppService,
		UserAppSvc: cfg.UserAppService,
		logger:     cfg.Logger,
	}

	// Create router.
	a.restContainer = restful.NewContainer()
	a.apiws = &restful.WebService{}
	a.apiws.Path(cfg.ServePefix).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	a.restContainer.Add(a.apiws)

	// Register routes.
	a.registerRoutes(cfg.ServePefix)

	// Enable cors.
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST"},
		CookiesAllowed: false,
		Container:      a.restContainer}
	a.restContainer.Filter(cors.Filter)
	a.restContainer.Filter(a.restContainer.OPTIONSFilter)

	return a, nil
}

func (a apiv1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.restContainer.ServeHTTP(w, r)
}
