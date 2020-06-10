package dice

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
)

// Service is the application service of dice logic.
type Service interface {
	// ListDiceTypes lists all the dice types supported by the app.
	ListDiceTypes(ctx context.Context) (*ListDiceTypesResponse, error)
	// CreateDiceRoll creates dice rolls.
	CreateDiceRoll(ctx context.Context, r CreateDiceRollRequest) (*CreateDiceRollResponse, error)
}

//go:generate mockery -case underscore -output dicemock -outpkg dicemock -name Service

// ServiceConfig is the service configuration.
type ServiceConfig struct {
	DiceRepository Repository
	Roller         Roller
	Logger         log.Logger
	IDGenerator    func() string
}

func (c *ServiceConfig) defaults() error {
	if c.DiceRepository == nil {
		return fmt.Errorf("dice.DiceRepository is required")
	}

	if c.Roller == nil {
		return fmt.Errorf("dice.Roller is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{"svc": "dice.Service"})

	if c.IDGenerator == nil {
		c.IDGenerator = func() string { return uuid.New().String() }
	}

	return nil
}

type service struct {
	diceRepository Repository
	roller         Roller
	logger         log.Logger
	idGen          func() string
}

// NewService returns a new dice.Service.
func NewService(cfg ServiceConfig) (Service, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return service{
		diceRepository: cfg.DiceRepository,
		roller:         cfg.Roller,
		logger:         cfg.Logger,
		idGen:          cfg.IDGenerator,
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

// CreateDiceRollRequest is the request of the roll.
type CreateDiceRollRequest struct {
	UserID string
	RoomID string
	Dice   []model.DieType
}

func (r CreateDiceRollRequest) validate() error {
	if r.RoomID == "" {
		return fmt.Errorf("config.RoomID is required")
	}

	if r.UserID == "" {
		return fmt.Errorf("config.UserID is required")
	}

	if len(r.Dice) == 0 {
		return fmt.Errorf("minimum config.Dice quantity is 1")
	}

	return nil
}

// CreateDiceRollResponse is the response to the RollDice request.
type CreateDiceRollResponse struct {
	DiceRoll model.DiceRoll
}

func (s service) CreateDiceRoll(ctx context.Context, r CreateDiceRollRequest) (*CreateDiceRollResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Create a dice roll.
	dice := []model.DieRoll{}
	for _, d := range r.Dice {
		dice = append(dice, model.DieRoll{
			ID:   s.idGen(),
			Type: d,
		})
	}

	dr := &model.DiceRoll{
		ID:   s.idGen(),
		Dice: dice,
	}

	// Roll'em all!
	err = s.roller.Roll(ctx, dr)
	if err != nil {
		return nil, fmt.Errorf("could not roll the dice: %w", err)
	}

	err = s.diceRepository.CreateDiceRoll(ctx, *dr)
	if err != nil {
		return nil, fmt.Errorf("could not store dice roll: %w", err)
	}

	// TODO(slok): Notify the roll.

	return &CreateDiceRollResponse{
		DiceRoll: *dr,
	}, nil
}
