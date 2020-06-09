package dice

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
)

var (
	// ErrNotValid will be used when something is not valid.
	ErrNotValid = fmt.Errorf("not valid")
)

// Service is the application service of dice logic.
type Service interface {
	// ListDiceTypes lists all the dice types supported by the app.
	ListDiceTypes(ctx context.Context) (*ListDiceTypesResponse, error)
}

//go:generate mockery -case underscore -output dicemock -outpkg dicemock -name Service

// ServiceConfig is the service configuration.
type ServiceConfig struct {
	Logger      log.Logger
	IDGenerator func() string
}

func (c *ServiceConfig) defaults() error {
	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	if c.IDGenerator == nil {
		c.IDGenerator = func() string { return uuid.New().String() }
	}

	return nil
}

type service struct {
	logger log.Logger
	idGen  func() string
}

// NewService returns a new dice.Service.
func NewService(cfg ServiceConfig) (Service, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return service{
		logger: cfg.Logger,
		idGen:  cfg.IDGenerator,
	}, nil
}

// ListDiceTypesResponse is the response for ListDiceTypes.
type ListDiceTypesResponse struct {
	DiceTypes []model.DieType
}

func (s service) ListDiceTypes(ctx context.Context) (*ListDiceTypesResponse, error) {
	return &ListDiceTypesResponse{
		DiceTypes: []model.DieType{
			model.DieTypeD4,
			model.DieTypeD6,
			model.DieTypeD8,
			model.DieTypeD10,
			model.DieTypeD12,
			model.DieTypeD20,
		},
	}, nil
}
