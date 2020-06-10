package room

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
	CreateRoom(ctx context.Context, r CreateRoomRequest) (*CreateRoomResponse, error)
}

//go:generate mockery -case underscore -output roommock -outpkg roommock -name Service

// ServiceConfig is the service configuration.
type ServiceConfig struct {
	RoomRepository Repository
	Logger         log.Logger
	IDGenerator    func() string
}

func (c *ServiceConfig) defaults() error {
	if c.RoomRepository == nil {
		return fmt.Errorf("config.RoomRepository is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{"svc": "room.Service"})

	if c.IDGenerator == nil {
		c.IDGenerator = func() string { return uuid.New().String() }
	}

	return nil
}

type service struct {
	roomRepo Repository
	logger   log.Logger
	idGen    func() string
}

// NewService returns a new room.Service.
func NewService(cfg ServiceConfig) (Service, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return service{
		roomRepo: cfg.RoomRepository,
		logger:   cfg.Logger,
		idGen:    cfg.IDGenerator,
	}, nil
}

// CreateRoomRequest is the request to CreateRoom.
type CreateRoomRequest struct {
	Name string
}

func (r CreateRoomRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("config.Name is required")
	}

	return nil
}

// CreateRoomResponse is the response to the CreateRoom request.
type CreateRoomResponse struct {
	Room model.Room
}

func (s service) CreateRoom(ctx context.Context, r CreateRoomRequest) (*CreateRoomResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotValid, err)
	}

	// Create a new room.
	room := model.Room{
		ID:   s.idGen(),
		Name: r.Name,
	}

	// Store room.
	err = s.roomRepo.CreateRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("could not store room: %w", err)
	}

	return &CreateRoomResponse{
		Room: room,
	}, nil
}
