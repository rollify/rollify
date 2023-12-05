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

// UserRepositoryConfig is the UserRepository configuration.
type UserRepositoryConfig struct {
	DBClient DBClient
	Table    string
	Logger   log.Logger
}

func (c *UserRepositoryConfig) defaults() error {
	if c.DBClient == nil {
		return fmt.Errorf("config.DBClient is required")
	}

	if c.Table == "" {
		c.Table = "user"
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	c.Logger = c.Logger.WithKV(log.KV{
		"repository":      "user",
		"repository-type": "mysql",
	})

	return nil
}

// UserRepository is a repository with MySQL implementation.
type UserRepository struct {
	db     DBClient
	table  string
	logger log.Logger
}

// NewUserRepository returns a new UserRepository.
func NewUserRepository(cfg UserRepositoryConfig) (*UserRepository, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &UserRepository{
		db:     cfg.DBClient,
		table:  cfg.Table,
		logger: cfg.Logger,
	}, nil
}

// CreateUser satisfies storage.UserRepository interface.
func (r *UserRepository) CreateUser(ctx context.Context, user model.User) error {
	// Map and create query.
	sqlUser := modelToSQLUser(user)
	query, args := userSQLBuilder.InsertInto(r.table, sqlUser).Build()

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

// ListRoomUsers satisfies storage.UserRepository interface.
func (r *UserRepository) ListRoomUsers(ctx context.Context, roomID string) (*storage.UserList, error) {
	sb := userSQLBuilder.SelectFrom(r.table)
	sb.Where(sb.Equal("room_id", roomID))
	query, args := sb.Build()

	// Get from database.
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("could not list users: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	su := &sqlUser{} // Reuse this, when mapping to model we will have a new instance.
	for rows.Next() {
		err := rows.Scan(userSQLBuilder.Addr(su)...)
		if err != nil {
			return nil, fmt.Errorf("could not scan SQL users: %w", err)
		}
		user := sqlToModelUser(su)
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("could not list users: %w", err)
	}

	return &storage.UserList{
		Items: users,
	}, nil
}

// UserExists storage.UserRepository interface.
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	// Create query.
	sb := userSQLBuilder.SelectFrom(r.table)
	sb.Where(sb.Equal("id", userID))
	query, args := sb.Build()

	// Get from database.
	row := r.db.QueryRowContext(ctx, query, args...)
	su := &sqlUser{}
	err := row.Scan(userSQLBuilder.Addr(su)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("missing user: %w: %w", internalerrors.ErrMissing, err)
		}

		return nil, fmt.Errorf("could not get user: %w", err)
	}

	// Map.
	user := sqlToModelUser(su)

	return &user, nil
}

// UserExists storage.UserRepository interface.
func (r *UserRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	// Create query.
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("*").From(r.table).Where(sb.Equal("id", userID))

	// Build and wrap for exists.
	b := sqlbuilder.Buildf("SELECT(EXISTS(%s))", sb)
	query, args := b.Build()

	// Get from database.
	row := r.db.QueryRowContext(ctx, query, args...)
	exists := false
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check user exists: %w", err)
	}

	return exists, nil
}

// UserExistsByNameInsensitive storage.UserRepository interface.
//
// We can directly check without any special query or transformation because:
// we use case insensitive collation for the specific name column:
// - Table is: CHARACTER SET utf8mb4 COLLATE utf8mb4_bin
// - Column `name` is: ... NOT NULL COLLATE utf8mb4_0900_ai_ci.
func (r *UserRepository) UserExistsByNameInsensitive(ctx context.Context, roomID, username string) (bool, error) {
	// Create query.
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("*").From(r.table).Where(
		sb.Equal("room_id", roomID),
		sb.Equal("name", username),
	)

	// Build and wrap for exists.
	b := sqlbuilder.Buildf("SELECT(EXISTS(%s))", sb)
	query, args := b.Build()

	// Get from database.
	row := r.db.QueryRowContext(ctx, query, args...)
	exists := false
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check user exists: %w", err)
	}

	return exists, nil
}

type sqlUser struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	RoomID    string    `db:"room_id"`
	CreatedAt time.Time `db:"created_at"`
}

func modelToSQLUser(r model.User) *sqlUser {
	return &sqlUser{
		ID:        r.ID,
		Name:      r.Name,
		RoomID:    r.RoomID,
		CreatedAt: r.CreatedAt,
	}
}

func sqlToModelUser(r *sqlUser) model.User {
	return model.User{
		ID:        r.ID,
		Name:      r.Name,
		RoomID:    r.RoomID,
		CreatedAt: r.CreatedAt,
	}
}

// Used as a light ORM by sqlbuilder.
var userSQLBuilder = sqlbuilder.NewStruct(&sqlUser{})

// Implementation assertions.
var _ storage.UserRepository = &UserRepository{}
