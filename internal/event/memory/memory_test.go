package memory_test

import (
	"context"
	"testing"

	"github.com/rollify/rollify/internal/event/memory"
	"github.com/rollify/rollify/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHubDiceRollCreatedEventsFlow tests all the hub flow form DiceRollCreated event
// this involves event notification and reception using via subscribing and unsubscribing.
func TestHubDiceRollCreatedEventsFlow(t *testing.T) {
	tests := map[string]struct {
		roomID       string
		userID       string
		unsubscribe  bool
		diceRolls    func() []model.DiceRoll
		expDiceRolls func() []model.DiceRoll
	}{
		"Having dice rolls on a room we are not subscribed, shouldn't receive the notifications.": {
			roomID: "room0-id",
			userID: "user0-id",
			diceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room2-id"},
				}
			},
			expDiceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{}
			},
		},

		"Having dice rolls on a room we are subscribed, shouldn receive the notifications.": {
			roomID: "room0-id",
			userID: "user0-id",
			diceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room0-id"},
				}
			},
			expDiceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room0-id"},
				}
			},
		},

		"Having a subscription on a room, we should receive only the notifications of that room.": {
			roomID: "room0-id",
			userID: "user0-id",
			diceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room0-id"},
					{ID: "dr1-id", RoomID: "room1-id"},
					{ID: "dr2-id", RoomID: "room2-id"},
					{ID: "dr3-id", RoomID: "room0-id"},
					{ID: "dr4-id", RoomID: "room1-id"},
					{ID: "dr5-id", RoomID: "room0-id"},
				}
			},
			expDiceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room0-id"},
					{ID: "dr3-id", RoomID: "room0-id"},
					{ID: "dr5-id", RoomID: "room0-id"},
				}
			},
		},

		"Having a subscription and then unsubscribing on a room, we shouldn't receive events.": {
			roomID:      "room0-id",
			userID:      "user0-id",
			unsubscribe: true,
			diceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{
					{ID: "dr0-id", RoomID: "room0-id"},
					{ID: "dr1-id", RoomID: "room1-id"},
					{ID: "dr2-id", RoomID: "room2-id"},
					{ID: "dr3-id", RoomID: "room0-id"},
					{ID: "dr4-id", RoomID: "room1-id"},
					{ID: "dr5-id", RoomID: "room0-id"},
				}
			},
			expDiceRolls: func() []model.DiceRoll {
				return []model.DiceRoll{}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			hub := memory.NewHub()

			// Subscribe with our check.
			gotDiceRolls := []model.DiceRoll{}
			err := hub.SubscribeDiceRollCreated(test.roomID, test.userID, func(_ context.Context, d model.DiceRoll) error {
				gotDiceRolls = append(gotDiceRolls, d)
				return nil
			})
			require.NoError(err)

			// In case we want to unsubscribe after subscription.
			if test.unsubscribe {
				err := hub.UnsubscribeDiceRollCreated(test.roomID, test.userID)
				require.NoError(err)
			}

			// Send
			for _, dr := range test.diceRolls() {
				err := hub.NotifyDiceRollCreated(context.TODO(), dr)
				require.NoError(err)
			}

			// TODO(slok): Wait for reception when making the implementation async.

			// Check.
			assert.Equal(test.expDiceRolls(), gotDiceRolls)
		})
	}
}
