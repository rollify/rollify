package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/rollify/rollify/internal/internalerrors"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
)

// DiceRollRepositoryConfig is the DiceRollRepository configuration.
type DiceRollRepositoryConfig struct {
	DBClient      DBClient
	DiceRollTable string
	DieRollTable  string
	Logger        log.Logger
}

func (c *DiceRollRepositoryConfig) defaults() error {
	if c.DBClient == nil {
		return fmt.Errorf("config.DBClient is required")
	}

	if c.DiceRollTable == "" {
		c.DiceRollTable = "dice_roll"
	}

	if c.DieRollTable == "" {
		c.DieRollTable = "die_roll"
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	c.Logger = c.Logger.WithKV(log.KV{
		"repository":      "diceRoll",
		"repository-type": "mysql",
	})

	return nil
}

// DiceRollRepository is a repository with MySQL implementation.
type DiceRollRepository struct {
	db            DBClient
	diceRollTable string
	dieRollTable  string
	logger        log.Logger
}

// NewDiceRollRepository returns a new DiceRollRepository.
func NewDiceRollRepository(cfg DiceRollRepositoryConfig) (*DiceRollRepository, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &DiceRollRepository{
		db:            cfg.DBClient,
		diceRollTable: cfg.DiceRollTable,
		dieRollTable:  cfg.DieRollTable,
		logger:        cfg.Logger,
	}, nil
}

// CreateDiceRoll satisfies storage.DiceRollRepository interface.
func (d DiceRollRepository) CreateDiceRoll(ctx context.Context, dr model.DiceRoll) error {
	// Dice roll insert query.
	sqlDiceRoll := modelToSQLDiceRoll(dr)
	diceRollQuery, diceRollArgs := insertDiceRollSQLBuilder.InsertInto(d.diceRollTable, sqlDiceRoll).Build()

	_, err := d.db.ExecContext(ctx, diceRollQuery, diceRollArgs...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", internalerrors.ErrAlreadyExists, err)
		}

		return fmt.Errorf("could not insert dice roll: %w", err)
	}

	// Prepare die rolls insert query
	sqlDieRolls := modelToSQLDieRolls(dr)
	dieRollQuery, dieRollArgs := dieRollSQLBuilder.InsertInto(d.dieRollTable, sqlDieRolls...).Build()

	_, err = d.db.ExecContext(ctx, dieRollQuery, dieRollArgs...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", internalerrors.ErrAlreadyExists, err)
		}

		return fmt.Errorf("could not insert die rolls: %w", err)
	}

	return nil
}

// ListDiceRolls satisfies storage.DiceRollRepository interface.
func (d DiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts storage.ListDiceRollsOpts) (*storage.DiceRollList, error) {
	return &storage.DiceRollList{}, nil
}

func modelToSQLDiceRoll(dr model.DiceRoll) *sqlInsertDiceRoll {
	return &sqlInsertDiceRoll{
		ID:        dr.ID,
		CreatedAt: dr.CreatedAt,
		RoomID:    dr.RoomID,
		UserID:    dr.UserID,
	}
}

// Returns []interface{} to make easier the insertion using sqlbuilder lib.
func modelToSQLDieRolls(dr model.DiceRoll) []interface{} {
	res := make([]interface{}, 0, len(dr.Dice))
	for _, d := range dr.Dice {
		res = append(res, &sqlDieRoll{
			ID:         d.ID,
			DiceRollID: dr.ID,
			DieTypeID:  d.Type.ID(),
			Side:       d.Side,
		})
	}
	return res
}

type sqlInsertDiceRoll struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	RoomID    string    `db:"room_id"`
	UserID    string    `db:"user_id"`
}

var insertDiceRollSQLBuilder = sqlbuilder.NewStruct(&sqlInsertDiceRoll{})

type sqlDiceRoll struct {
	sqlInsertDiceRoll
	Serial uint64 `db:"serial"`
}

var diceRollSQLBuilder = sqlbuilder.NewStruct(&sqlDiceRoll{})

type sqlDieRoll struct {
	ID         string `db:"id"`
	DiceRollID string `db:"dice_roll_id"`
	DieTypeID  string `db:"die_type_id"`
	Side       uint   `db:"side"`
}

var dieRollSQLBuilder = sqlbuilder.NewStruct(&sqlDieRoll{})

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
