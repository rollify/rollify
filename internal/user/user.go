package user

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// Service is the application service of users logic.
type Service interface {
	// Creates an user for a specific room.
	CreateUser(ctx context.Context, r CreateUserRequest) (*CreateUserResponse, error)
	// Lists users for a specific room.
	ListUsers(ctx context.Context, r ListUsersRequest) (*ListUsersResponse, error)
	// Get an user by its ID.
	GetUser(ctx context.Context, r GetUserRequest) (*GetUserResponse, error)
}

//go:generate mockery --case underscore --output usermock --outpkg usermock --name Service

// ServiceConfig is the service configuration.
type ServiceConfig struct {
	UserRepository storage.UserRepository
	RoomRepository storage.RoomRepository
	Logger         log.Logger
	IDGenerator    func() string
	TimeNowFunc    func() time.Time
}

func (c *ServiceConfig) defaults() error {
	if c.UserRepository == nil {
		return fmt.Errorf("config.UserRepository is required")
	}

	if c.RoomRepository == nil {
		return fmt.Errorf("config.RoomRepository is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}
	c.Logger = c.Logger.WithKV(log.KV{"svc": "user.Service"})

	if c.IDGenerator == nil {
		c.IDGenerator = func() string { return uuid.New().String() }
	}

	if c.TimeNowFunc == nil {
		c.TimeNowFunc = time.Now
	}

	return nil
}

type service struct {
	userRepo storage.UserRepository
	roomRepo storage.RoomRepository
	logger   log.Logger
	idGen    func() string
	timeNow  func() time.Time
}

// NewService returns a new user.Service.
func NewService(cfg ServiceConfig) (Service, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return service{
		userRepo: cfg.UserRepository,
		roomRepo: cfg.RoomRepository,
		logger:   cfg.Logger,
		idGen:    cfg.IDGenerator,
		timeNow:  cfg.TimeNowFunc,
	}, nil
}

// CreateUserRequest is the request to CreateUser.
type CreateUserRequest struct {
	Name   string
	RoomID string
}

var userNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _\-.']+$`)

func (r CreateUserRequest) validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.RoomID == "" {
		return fmt.Errorf("roomID is required")
	}

	if !userNameRegex.MatchString(r.Name) {
		return fmt.Errorf("name regex is not valid, must be %s", userNameRegex.String())
	}
	return nil
}

// CreateUserResponse is the response to the CreateUser request.
type CreateUserResponse struct {
	User model.User
}

func (s service) CreateUser(ctx context.Context, r CreateUserRequest) (*CreateUserResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Check the room exists.
	exists, err := s.roomRepo.RoomExists(ctx, r.RoomID)
	if err != nil {
		return nil, fmt.Errorf("could not check room exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("room does not exist: %w", internalerrors.ErrNotValid)
	}

	// Check user exists by room and being case insensitive.
	exists, err = s.userRepo.UserExistsByNameInsensitive(ctx, r.RoomID, r.Name)
	if err != nil {
		return nil, fmt.Errorf("could not check user exists: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user already exists: %w", internalerrors.ErrAlreadyExists)
	}

	// Create a new user.
	user := model.User{
		ID:        s.idGen(),
		CreatedAt: s.timeNow().UTC(),
		RoomID:    r.RoomID,
		Name:      r.Name,
	}

	// Store user.
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("could not store user: %w", err)
	}

	return &CreateUserResponse{
		User: user,
	}, nil
}

// ListUsersRequest is the request to ListUsers.
type ListUsersRequest struct {
	RoomID string
}

func (r ListUsersRequest) validate() error {
	if r.RoomID == "" {
		return fmt.Errorf("roomID is required")
	}

	return nil
}

// ListUsersResponse is the response to the ListUsers request.
type ListUsersResponse struct {
	Users []model.User
}

func (s service) ListUsers(ctx context.Context, r ListUsersRequest) (*ListUsersResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	exists, err := s.roomRepo.RoomExists(ctx, r.RoomID)
	if err != nil {
		return nil, fmt.Errorf("could not check if room exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("%w: room does not exist", internalerrors.ErrNotValid)
	}

	us, err := s.userRepo.ListRoomUsers(ctx, r.RoomID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve users: %w", err)
	}

	return &ListUsersResponse{
		Users: us.Items,
	}, nil
}

// GetUserRequest is the request to GetUser.
type GetUserRequest struct {
	UserID string
}

func (r GetUserRequest) validate() error {
	if r.UserID == "" {
		return fmt.Errorf("userID is required")
	}

	return nil
}

// GetUserResponse is the response to the GetUser request.
type GetUserResponse struct {
	User model.User
}

func (s service) GetUser(ctx context.Context, r GetUserRequest) (*GetUserResponse, error) {
	err := r.validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", internalerrors.ErrNotValid, err)
	}

	// Get the User.
	user, err := s.userRepo.GetUserByID(ctx, r.UserID)
	if err != nil {
		return nil, fmt.Errorf("could not get the user: %w", err)
	}

	return &GetUserResponse{
		User: *user,
	}, nil
}
