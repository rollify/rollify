package dice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/dice/dicemock"
	"github.com/rollify/rollify/internal/event/eventmock"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/storage/storagemock"
)

func TestServiceListDiceTypes(t *testing.T) {
	tests := map[string]struct {
		config  dice.ServiceConfig
		expResp func() *dice.ListDiceTypesResponse
		expErr  bool
	}{
		"Listing dice types should return all the available dice types.": {
			expResp: func() *dice.ListDiceTypesResponse {
				return &dice.ListDiceTypesResponse{
					DiceTypes: []model.DieType{
						model.DieTypeD4,
						model.DieTypeD6,
						model.DieTypeD8,
						model.DieTypeD10,
						model.DieTypeD12,
						model.DieTypeD20,
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			test.config.Roller = &dicemock.Roller{}
			test.config.DiceRollRepository = &storagemock.DiceRollRepository{}
			test.config.RoomRepository = &storagemock.RoomRepository{}
			test.config.UserRepository = &storagemock.UserRepository{}
			test.config.EventNotifier = &eventmock.Notifier{}
			svc, err := dice.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.ListDiceTypes(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceCreateDiceRoll(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		config  dice.ServiceConfig
		mock    func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier)
		req     func() dice.CreateDiceRollRequest
		expResp func() *dice.CreateDiceRollResponse
		expErr  bool
	}{
		"Having a dice roll request without room should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request without user should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "room-id",
					UserID: "",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request without dice should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "room-id",
					UserID: "test-id",
					Dice:   []model.DieType{},
				}
			},
			expErr: true,
		},

		"Having a dice roll request with too much dice should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
			},
			req: func() dice.CreateDiceRollRequest {
				ds := []model.DieType{}
				for i := 0; i < 101; i++ {
					ds = append(ds, model.DieTypeD10)
				}
				return dice.CreateDiceRollRequest{
					RoomID: "room-id",
					UserID: "test-id",
					Dice:   ds,
				}
			},
			expErr: true,
		},

		"Having a dice roll request with a room that does not exists it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, "test-room").Once().Return(false, nil)
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request if checking if the room exists fail, it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, "test-room").Once().Return(false, fmt.Errorf("wanted error"))
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request with a user that does not exists it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, "user-id").Once().Return(false, nil)
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request if checking if the user exists fail, it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, "user-id").Once().Return(false, fmt.Errorf("wanted error"))
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request it should create a dice roll, roll them, store and notify.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				// Expexted dice roll call.
				exp := &model.DiceRoll{
					ID:        "test",
					CreatedAt: t0,
					RoomID:    "test-room",
					UserID:    "user-id",
					Dice: []model.DieRoll{
						{ID: "test", Type: model.DieTypeD6},
						{ID: "test", Type: model.DieTypeD8},
						{ID: "test", Type: model.DieTypeD10},
					},
				}
				roomRepo.On("RoomExists", mock.Anything, "test-room").Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, "user-id").Once().Return(true, nil)
				roller.On("Roll", mock.Anything, exp).Once().Return(nil)
				diceRollRepo.On("CreateDiceRoll", mock.Anything, *exp).Once().Return(nil)
				notifier.On("NotifyDiceRollCreated", mock.Anything, *exp).Once().Return(nil)
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice: []model.DieType{
						model.DieTypeD6,
						model.DieTypeD8,
						model.DieTypeD10,
					},
				}
			},
			expResp: func() *dice.CreateDiceRollResponse {
				return &dice.CreateDiceRollResponse{
					DiceRoll: model.DiceRoll{
						ID:        "test",
						CreatedAt: t0,
						RoomID:    "test-room",
						UserID:    "user-id",
						Dice: []model.DieRoll{
							{ID: "test", Type: model.DieTypeD6},
							{ID: "test", Type: model.DieTypeD8},
							{ID: "test", Type: model.DieTypeD10},
						},
					},
				}
			},
		},

		"Having a dice roll request and failing the dice roll process, it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				roller.On("Roll", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("wanted error"))
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request if storage fails, it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				roller.On("Roll", mock.Anything, mock.Anything).Once().Return(nil)
				diceRollRepo.On("CreateDiceRoll", mock.Anything, mock.Anything).Once().Return(errors.New("wanted error"))
			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},

		"Having a dice roll request if notification fails, it should fail.": {
			mock: func(roller *dicemock.Roller, diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository, userRepo *storagemock.UserRepository, notifier *eventmock.Notifier) {
				roomRepo.On("RoomExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				userRepo.On("UserExists", mock.Anything, mock.Anything).Once().Return(true, nil)
				roller.On("Roll", mock.Anything, mock.Anything).Once().Return(nil)
				diceRollRepo.On("CreateDiceRoll", mock.Anything, mock.Anything).Once().Return(nil)
				notifier.On("NotifyDiceRollCreated", mock.Anything, mock.Anything).Once().Return(errors.New("wanted error"))

			},
			req: func() dice.CreateDiceRollRequest {
				return dice.CreateDiceRollRequest{
					RoomID: "test-room",
					UserID: "user-id",
					Dice:   []model.DieType{model.DieTypeD6},
				}
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mrol := &dicemock.Roller{}
			mdrrep := &storagemock.DiceRollRepository{}
			mrrep := &storagemock.RoomRepository{}
			murep := &storagemock.UserRepository{}
			mevn := &eventmock.Notifier{}
			test.mock(mrol, mdrrep, mrrep, murep, mevn)

			test.config.Roller = mrol
			test.config.DiceRollRepository = mdrrep
			test.config.RoomRepository = mrrep
			test.config.UserRepository = murep
			test.config.EventNotifier = mevn
			test.config.IDGenerator = func() string { return "test" }
			test.config.TimeNowFunc = func() time.Time { return t0 }

			svc, err := dice.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.CreateDiceRoll(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}

func TestServiceListDiceRolls(t *testing.T) {
	tests := map[string]struct {
		config  dice.ServiceConfig
		mock    func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository)
		req     func() dice.ListDiceRollsRequest
		expResp func() *dice.ListDiceRollsResponse
		expErr  bool
	}{
		"Having a list dice roll request without room should fail.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "",
					UserID: "user-id",
				}
			},
			expErr: true,
		},

		"Having a list dice roll request with an error listing dice rolls, should fail.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
				diceRollRepo.On("ListDiceRolls", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("wanted error"))
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
				}
			},
			expErr: true,
		},

		"Having a list dice roll request should list dice rolls.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
				expOpts := storage.ListDiceRollsOpts{
					RoomID: "room-id",
					UserID: "user-id",
				}
				dr := &storage.DiceRollList{
					Items: []model.DiceRoll{
						{ID: "dr1"},
						{ID: "dr2"},
					},
				}
				diceRollRepo.On("ListDiceRolls", mock.Anything, mock.Anything, expOpts).Once().Return(dr, nil)
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
				}
			},
			expResp: func() *dice.ListDiceRollsResponse {
				return &dice.ListDiceRollsResponse{
					DiceRolls: []model.DiceRoll{
						{ID: "dr1"},
						{ID: "dr2"},
					},
				}
			},
		},

		"Not having pagination should set safe defaults.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
				expPageOpts := model.PaginationOpts{
					Order: model.PaginationOrderDesc,
					Size:  100,
				}
				dr := &storage.DiceRollList{}
				diceRollRepo.On("ListDiceRolls", mock.Anything, expPageOpts, mock.Anything).Once().Return(dr, nil)
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
				}
			},
			expResp: func() *dice.ListDiceRollsResponse {
				return &dice.ListDiceRollsResponse{}
			},
		},

		"Having custom pagination should use it.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
				expPageOpts := model.PaginationOpts{
					Cursor: "threepwood",
					Order:  model.PaginationOrderAsc,
					Size:   93,
				}
				dr := &storage.DiceRollList{}
				diceRollRepo.On("ListDiceRolls", mock.Anything, expPageOpts, mock.Anything).Once().Return(dr, nil)
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
					PageOpts: model.PaginationOpts{
						Cursor: "threepwood",
						Order:  model.PaginationOrderAsc,
						Size:   93,
					},
				}
			},
			expResp: func() *dice.ListDiceRollsResponse {
				return &dice.ListDiceRollsResponse{}
			},
		},

		"Having a pagination return from the repository, should be returned.": {
			mock: func(diceRollRepo *storagemock.DiceRollRepository, roomRepo *storagemock.RoomRepository) {
				dr := &storage.DiceRollList{
					Cursors: model.PaginationCursors{
						FirstCursor: "first",
						LastCursor:  "second",
						HasNext:     true,
						HasPrevious: true,
					},
				}
				diceRollRepo.On("ListDiceRolls", mock.Anything, mock.Anything, mock.Anything).Once().Return(dr, nil)
			},
			req: func() dice.ListDiceRollsRequest {
				return dice.ListDiceRollsRequest{
					RoomID: "room-id",
					UserID: "user-id",
					PageOpts: model.PaginationOpts{
						Cursor: "threepwood",
						Order:  model.PaginationOrderAsc,
						Size:   93,
					},
				}
			},
			expResp: func() *dice.ListDiceRollsResponse {
				return &dice.ListDiceRollsResponse{
					Cursors: model.PaginationCursors{
						FirstCursor: "first",
						LastCursor:  "second",
						HasNext:     true,
						HasPrevious: true,
					},
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks
			mdrrep := &storagemock.DiceRollRepository{}
			mrrep := &storagemock.RoomRepository{}
			test.mock(mdrrep, mrrep)

			test.config.Roller = &dicemock.Roller{}
			test.config.DiceRollRepository = mdrrep
			test.config.RoomRepository = mrrep
			test.config.UserRepository = &storagemock.UserRepository{}
			test.config.EventNotifier = &eventmock.Notifier{}
			test.config.IDGenerator = func() string { return "test" }

			svc, err := dice.NewService(test.config)
			require.NoError(err)

			gotResp, err := svc.ListDiceRolls(context.TODO(), test.req())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp(), gotResp)
			}
		})
	}
}
