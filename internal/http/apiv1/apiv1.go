package apiv1

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/log"
)

// Config is the configuration to serve the API.
type Config struct {
	DiceAppService dice.Service
	ServePefix     string
	Logger         log.Logger
}

func (c *Config) defaults() error {
	if c.DiceAppService == nil {
		return fmt.Errorf("dice.Service application service is required")
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
		logger:     cfg.Logger,
	}

	// Create router.
	a.restContainer = restful.NewContainer()
	a.apiws = &restful.WebService{}
	a.apiws.Path(cfg.ServePefix).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	a.restContainer.Add(a.apiws)

	a.registerRoutes(cfg.ServePefix)

	return a, nil
}

func (a apiv1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.restContainer.ServeHTTP(w, r)
}
