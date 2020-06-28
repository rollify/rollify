package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"

	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// RoomRepositoryConfig is the RoomRepository configuration.
type RoomRepositoryConfig struct {
	DBClient DBClient
	Table    string
	Logger   log.Logger
}

func (c *RoomRepositoryConfig) defaults() error {
	if c.DBClient == nil {
		return fmt.Errorf("config.DBClient is required")
	}

	if c.Table == "" {
		c.Table = "room"
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	c.Logger = c.Logger.WithKV(log.KV{
		"repository":      "room",
		"repository-type": "mysql",
	})

	return nil
}

// RoomRepository is a repository with MySQL implementation.
type RoomRepository struct {
	db     DBClient
	table  string
	logger log.Logger
}

// NewRoomRepository returns a new RoomRepository.
func NewRoomRepository(cfg RoomRepositoryConfig) (*RoomRepository, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &RoomRepository{
		db:     cfg.DBClient,
		table:  cfg.Table,
		logger: cfg.Logger,
	}, nil
}

// CreateRoom satisfies storage.RoomRepository interface.
func (r *RoomRepository) CreateRoom(ctx context.Context, room model.Room) error {
	// Map and create query.
	sqlRoom := modelToSQLRoom(room)
	query, args := roomSQLBuilder.InsertInto(r.table, sqlRoom).Build()

	// Insert in database.
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", internalerrors.ErrAlreadyExists, err)
		}

		return err
	}

	return nil
}

// GetRoom satisfies storage.RoomRepository interface.
func (r *RoomRepository) GetRoom(ctx context.Context, id string) (*model.Room, error) {
	// Create query.
	sb := roomSQLBuilder.SelectFrom(r.table)
	sb.Where(sb.Equal("id", id))
	query, args := sb.Build()

	// Get from database.
	row := r.db.QueryRowContext(ctx, query, args...)
	sr := &sqlRoom{}
	err := row.Scan(roomSQLBuilder.Addr(sr)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("missing room: %w: %e", internalerrors.ErrMissing, err)
		}

		return nil, fmt.Errorf("could not get room: %w", err)
	}

	// Map.
	room := sqlRoomToModel(sr)

	return room, nil
}

// RoomExists satisfies storage.RoomRepository interface.
func (r *RoomRepository) RoomExists(ctx context.Context, id string) (bool, error) {
	// Create query.
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("*").From(r.table).Where(sb.Equal("id", id))

	// Build and wrap for exists.
	b := sqlbuilder.Buildf("SELECT(EXISTS(%s))", sb)
	query, args := b.Build()

	// Get from database.
	row := r.db.QueryRowContext(ctx, query, args...)
	exists := false
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check room exists: %w", err)
	}

	return exists, nil
}

type sqlRoom struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func modelToSQLRoom(r model.Room) *sqlRoom {
	return &sqlRoom{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}

func sqlRoomToModel(r *sqlRoom) *model.Room {
	return &model.Room{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}

// Used as a light ORM by sqlbuilder.
var roomSQLBuilder = sqlbuilder.NewStruct(&sqlRoom{})

// Implementation assertions.
var _ storage.RoomRepository = &RoomRepository{}
