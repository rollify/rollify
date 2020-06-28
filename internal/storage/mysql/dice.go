package mysql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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
// TODO(slok): Would be interesting to add this in transaction mode? for now the simplest way.
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

	// Prepare die rolls insert query.
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

type cursor struct {
	Serial int `json:"serial"`
}

// ListDiceRolls satisfies storage.DiceRollRepository interface.
func (d DiceRollRepository) ListDiceRolls(ctx context.Context, pageOpts model.PaginationOpts, filterOpts storage.ListDiceRollsOpts) (*storage.DiceRollList, error) {
	// We want something similar to this query:
	//
	// SELECT drs.id, drs.created_at, drs.room_id, drs.user_id, drs.serial, dr.id, dr.die_type_id, dr.side
	// FROM die_roll dr
	// JOIN (
	//     SELECT id, created_at, room_id, user_id, serial
	//	       FROM dice_roll
	//  	   WHERE room_id = "f72bebf6-506b-40d3-9772-653204174515"
	//		   AND serial > 123
	//     ORDER BY serial ASC
	//     LIMIT 100
	// ) AS drs ON dr.dice_roll_id = drs.id
	// ORDER BY serial ASC

	sb := sqlbuilder.NewSelectBuilder()
	joinSb := sqlbuilder.NewSelectBuilder()

	sb.Select("drs.id", "drs.created_at", "drs.room_id", "drs.user_id", "drs.serial", "dr.id", "dr.die_type_id", "dr.side").
		From(d.dieRollTable+" dr").
		Join(sb.BuilderAs(joinSb, "drs"), "dr.dice_roll_id = drs.id")

	joinSb.Select("id", "created_at", "room_id", "user_id", "serial").
		From(d.diceRollTable).
		Where(joinSb.Equal("room_id", filterOpts.RoomID))

	// If limit set.
	if pageOpts.Size > 0 {
		joinSb.Limit(int(pageOpts.Size))
	}

	// In case we need to filter also by user.
	if filterOpts.UserID != "" {
		joinSb.Where(joinSb.Equal("drs.user_id", filterOpts.UserID))
	}

	// Add order.
	if pageOpts.Order == model.PaginationOrderAsc {
		joinSb.OrderBy("serial ASC")
		sb.OrderBy("serial ASC")
	} else {
		joinSb.OrderBy("serial DESC")
		sb.OrderBy("serial DESC")
	}

	// In case of cursor select from there.
	if pageOpts.Cursor != "" {
		cr, err := strToCursor(pageOpts.Cursor)
		if err != nil {
			return nil, fmt.Errorf("could not get information from cursor: %w", err)
		}

		// Add cursor based filtering.
		if pageOpts.Order == model.PaginationOrderAsc {
			joinSb.Where(joinSb.GreaterThan("serial", cr.Serial))
		} else {
			joinSb.Where(joinSb.LessThan("serial", cr.Serial))
		}
	}

	// Get from database.
	query, args := sb.Build()
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("could not list dice rolls: %w", err)
	}
	defer rows.Close()

	// We use resultDiceRolls to maintain the order and the index to get the dice roll to set the die rolls.
	resultDiceRolls := []*model.DiceRoll{}
	indexDiceRolls := map[string]*model.DiceRoll{}
	drs := &sqlDiceRoll{} // Reuse this, when mapping to model we will have a new instance.
	dr := &sqlDieRoll{}   // Reuse this, when mapping to model we will have a new instance.
	for rows.Next() {
		err := rows.Scan(&drs.ID, &drs.CreatedAt, &drs.RoomID, &drs.UserID, &drs.Serial, &dr.ID, &dr.DieTypeID, &dr.Side)
		if err != nil {
			return nil, fmt.Errorf("could not scan SQL dice rolls: %w", err)
		}

		// Get or create the dice roll.
		diceRoll, ok := indexDiceRolls[drs.ID]
		if !ok {
			diceRoll = sqlToModelDiceRoll(drs)
			indexDiceRolls[drs.ID] = diceRoll
			resultDiceRolls = append(resultDiceRolls, diceRoll)
		}

		// Add the die rolls.
		dieroll, err := sqlToModelDieRoll(dr)
		if err != nil {
			return nil, fmt.Errorf("could not map SQL die roll to model: %w", err)
		}
		diceRoll.Dice = append(diceRoll.Dice, *dieroll)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("could not list dice rolls: %w", err)
	}

	// Create cursors.
	firstCursor := ""
	lastCursor := ""
	if len(resultDiceRolls) > 0 {
		firstCursor, err = serialToCursorStr(resultDiceRolls[0].Serial)
		if err != nil {
			return nil, fmt.Errorf("could not marshal cursor: %w", err)
		}

		lastCursor, err = serialToCursorStr(resultDiceRolls[len(resultDiceRolls)-1].Serial)
		if err != nil {
			return nil, fmt.Errorf("could not marshal cursor: %w", err)
		}
	}

	items := make([]model.DiceRoll, 0, len(resultDiceRolls))
	for _, dr := range resultDiceRolls {
		items = append(items, *dr)
	}

	return &storage.DiceRollList{
		Items: items,
		Cursors: model.PaginationCursors{
			FirstCursor: firstCursor,
			LastCursor:  lastCursor,
			// If we have cursor, then we always have previous.
			HasPrevious: pageOpts.Cursor != "",
			// If the max size is the number of items, we have high probability of having more, if not, on next queyr, it wony.
			HasNext: len(items) >= int(pageOpts.Size),
		},
	}, nil
}

func strToCursor(s string) (*cursor, error) {
	c, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 cursor: %w: %s", internalerrors.ErrNotValid, err)
	}

	cr := &cursor{}
	err = json.Unmarshal(c, cr)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json cursor: %w: %s", internalerrors.ErrNotValid, err)
	}

	return cr, nil
}

func serialToCursorStr(s uint) (string, error) {
	c := cursor{Serial: int(s)}
	jc, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("could not marshal cursor: %w", err)
	}

	cs := base64.StdEncoding.EncodeToString([]byte(jc))
	return cs, nil
}

func modelToSQLDiceRoll(dr model.DiceRoll) *sqlInsertDiceRoll {
	return &sqlInsertDiceRoll{
		ID:        dr.ID,
		CreatedAt: dr.CreatedAt,
		RoomID:    dr.RoomID,
		UserID:    dr.UserID,
	}
}

func sqlToModelDiceRoll(dr *sqlDiceRoll) *model.DiceRoll {
	return &model.DiceRoll{
		ID:        dr.ID,
		Serial:    uint(dr.Serial),
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

func sqlToModelDieRoll(dr *sqlDieRoll) (*model.DieRoll, error) {
	dt, ok := model.DiceTypes[dr.DieTypeID]
	if !ok {
		return nil, fmt.Errorf("invalid dice type: %s", dr.DieTypeID)
	}

	return &model.DieRoll{
		ID:   dr.ID,
		Type: dt,
		Side: dr.Side,
	}, nil
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

type sqlDieRoll struct {
	ID         string `db:"id"`
	DiceRollID string `db:"dice_roll_id"`
	DieTypeID  string `db:"die_type_id"`
	Side       uint   `db:"side"`
}

var dieRollSQLBuilder = sqlbuilder.NewStruct(&sqlDieRoll{})

// Implementation assertions.
var _ storage.DiceRollRepository = &DiceRollRepository{}
