package dice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rollify/rollify/internal/event"
	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// Service is the application service of dice logic.
type Service interface {
	// ListDiceTypes lists all the dice types supported by the app.
	ListDiceTypes(ctx context.Context) (*ListDiceTypesResponse, error)
	// CreateDiceRoll creates dice rolls.
	CreateDiceRoll(ctx context.Context, r CreateDiceRollRequest) (*CreateDiceRollResponse, error)
	// ListDiceRolls lists dice rolls.
	ListDiceRolls(ctx context.Context, r ListDiceRollsRequest) (*ListDiceRollsResponse, error)
	// Subscribes to a diceroll created events
	SubscribeDiceRollCreated(ctx context.Context, r SubscribeDiceRollCreatedRequest) (*SubscribeDiceRollCreatedResponse, error)
}

//go:generate mockery -case underscore -output dicemock -outpkg dicemock -name Service

// ServiceConfig is the service configuration.
type ServiceConfig struct {
	DiceRollRepository storage.DiceRollRepository
	RoomRepository     storage.RoomRepository
	UserRepository     storage.UserRepository
	Roller             Roller
	EventNotifier      event.Notifier
	EventSubscriber    event.Subscriber
	Logger             log.Logger
	IDGenerator        func() string
	TimeNowFunc        func() time.Time
}

func (c *ServiceConfig) defaults() error {
	if c.DiceRollRepository == nil {
		return fmt.Errorf("storage.DiceRollRepository is required")
	}

	if c.RoomRepository == nil {
		return fmt.Errorf("storage.RoomRepository is required")
	}

	if c.UserRepository == nil {
		return fmt.Errorf("storage.UserRepository is required")
	}

	if c.Roller == nil {
		return fmt.Errorf("dice.Roller is required")
	}

	if c.EventNotifier == nil {
		return fmt.Errorf("event.Notifier is required")
	}

	if c.EventSubscriber == nil {
		return fmt.Errorf("event.Subscriber is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{"svc": "dice.Service"})

	if c.IDGenerator == nil {
		c.IDGenerator = func() string { return uuid.New().String() }
	}

	if c.TimeNowFunc == nil {
		c.TimeNowFunc = time.Now
	}

	return nil
}

type service struct {
	diceRollRepository storage.DiceRollRepository
	roomRepository     storage.RoomRepository
	userRepository     storage.UserRepository
	roller             Roller
	eventNotifier      event.Notifier
	eventSubscriber    event.Subscriber
	logger             log.Logger
	idGen              func() string
	timeNow            func() time.Time
}

// NewService returns a new dice.Service.
func NewService(cfg ServiceConfig) (Service, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return service{
		diceRollRepository: cfg.DiceRollRepository,
		roomRepository:     cfg.RoomRepository,
		userRepository:     cfg.UserRepository,
		roller:             cfg.Roller,
		eventNotifier:      cfg.EventNotifier,
		eventSubscriber:    cfg.EventSubscriber,
		logger:             cfg.Logger,
		idGen:              cfg.IDGenerator,
		timeNow:            cfg.TimeNowFunc,
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

// CreateDiceRollRequest is the request for CreateDiceRoll.
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

	if len(r.Dice) > 100 {
		return fmt.Errorf("max config.Dice quantity is 100, got %d", len(r.Dice))
	}

	return nil
}

// CreateDiceRollResponse is the response for CreateDiceRoll.
type CreateDiceRollResponse struct {
	DiceRoll model.DiceRoll
}

func (s service) CreateDiceRoll(ctx context.Context, r CreateDiceRollRequest) (*CreateDiceRollResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Check the room exists.
	roomExists, err := s.roomRepository.RoomExists(ctx, r.RoomID)
	if err != nil {
		return nil, fmt.Errorf("could not check if room exists: %w", err)
	}
	if !roomExists {
		return nil, fmt.Errorf("room does not exists: %w", internalerrors.ErrNotValid)
	}

	// Check the user exists.
	userExist, err := s.userRepository.UserExists(ctx, r.UserID)
	if err != nil {
		return nil, fmt.Errorf("could not check if user exists: %w", err)
	}
	if !userExist {
		return nil, fmt.Errorf("user does not exists: %w", internalerrors.ErrNotValid)
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
		ID:        s.idGen(),
		CreatedAt: s.timeNow().UTC(),
		RoomID:    r.RoomID,
		UserID:    r.UserID,
		Dice:      dice,
	}

	// Roll'em all!
	err = s.roller.Roll(ctx, dr)
	if err != nil {
		return nil, fmt.Errorf("could not roll the dice: %w", err)
	}

	// Store the dice roll.
	err = s.diceRollRepository.CreateDiceRoll(ctx, *dr)
	if err != nil {
		return nil, fmt.Errorf("could not store dice roll: %w", err)
	}

	// Send dice roll event.
	// TODO(slok): should this be returned as an error?.
	ev := model.EventDiceRollCreated{DiceRoll: *dr}
	err = s.eventNotifier.NotifyDiceRollCreated(ctx, ev)
	if err != nil {
		return nil, fmt.Errorf("could not send dice roll created event: %w", err)
	}

	return &CreateDiceRollResponse{
		DiceRoll: *dr,
	}, nil
}

// ListDiceRollsRequest is the request for ListDiceRolls.
type ListDiceRollsRequest struct {
	UserID   string
	RoomID   string
	PageOpts model.PaginationOpts
}

func (r ListDiceRollsRequest) validate() error {
	if r.RoomID == "" {
		return fmt.Errorf("config.RoomID is required")
	}

	return nil
}

// ListDiceRollsResponse is the response for ListDiceRolls.
type ListDiceRollsResponse struct {
	DiceRolls []model.DiceRoll
	Cursors   model.PaginationCursors
}

func (s service) ListDiceRolls(ctx context.Context, r ListDiceRollsRequest) (*ListDiceRollsResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Set up pagination defaults.
	if r.PageOpts.Size <= 0 || r.PageOpts.Size > 100 {
		r.PageOpts.Size = 100
	}
	if r.PageOpts.Order == model.PaginationOrderDefault {
		r.PageOpts.Order = model.PaginationOrderDesc
	}

	drs, err := s.diceRollRepository.ListDiceRolls(ctx, r.PageOpts, storage.ListDiceRollsOpts{
		RoomID: r.RoomID,
		UserID: r.UserID,
	})

	if err != nil {
		return nil, fmt.Errorf("could not get dice roll list: %w", err)
	}

	return &ListDiceRollsResponse{
		DiceRolls: drs.Items,
		Cursors:   drs.Cursors,
	}, nil
}

// SubscribeDiceRollCreatedRequest is the request for SubscribeDiceRollCreated.
type SubscribeDiceRollCreatedRequest struct {
	RoomID       string
	EventHandler func(context.Context, model.EventDiceRollCreated) error
}

func (r SubscribeDiceRollCreatedRequest) validate() error {
	if r.RoomID == "" {
		return fmt.Errorf("roomID is required")
	}

	if r.EventHandler == nil {
		return fmt.Errorf("eventHandler is required")
	}

	return nil
}

// SubscribeDiceRollCreatedResponse is the response for SubscribeDiceRollCreated.
type SubscribeDiceRollCreatedResponse struct {
	UnsubscribeFunc func() error
}

func (s service) SubscribeDiceRollCreated(ctx context.Context, r SubscribeDiceRollCreatedRequest) (*SubscribeDiceRollCreatedResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Check the room exists.
	roomExists, err := s.roomRepository.RoomExists(ctx, r.RoomID)
	if err != nil {
		return nil, fmt.Errorf("could not check if room exists: %w", err)
	}
	if !roomExists {
		return nil, fmt.Errorf("room does not exists: %w", internalerrors.ErrNotValid)
	}

	// Create a subscription ID and subscribe.
	subscriptionID := s.idGen()
	err = s.eventSubscriber.SubscribeDiceRollCreated(subscriptionID, r.RoomID, r.EventHandler)
	if err != nil {
		return nil, fmt.Errorf("could not subscribe to diceRollCreated events: %w", err)
	}

	return &SubscribeDiceRollCreatedResponse{
		UnsubscribeFunc: func() error {
			return s.eventSubscriber.UnsubscribeDiceRollCreated(subscriptionID, r.RoomID)
		},
	}, nil

}
