package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rollify/rollify/internal/event/memory"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
)

// TestHubDiceRollCreatedEventsFlow tests all the hub flow form DiceRollCreated event
// this involves event notification and reception using via subscribing and unsubscribing.
func TestHubDiceRollCreatedEventsFlow(t *testing.T) {
	tests := map[string]struct {
		roomID      string
		id          string
		unsubscribe bool
		events      func() []model.EventDiceRollCreated
		expEvents   func() []model.EventDiceRollCreated
	}{
		"Having dice rolls on a room we are not subscribed, shouldn't receive the notifications.": {
			roomID: "room0-id",
			id:     "user0-id",
			events: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room2-id"}},
				}
			},
			expEvents: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{}
			},
		},

		"Having dice rolls on a room we are subscribed, shouldn receive the notifications.": {
			roomID: "room0-id",
			id:     "user0-id",
			events: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id"}},
				}
			},
			expEvents: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id"}},
				}
			},
		},

		"Having a subscription on a room, we should receive only the notifications of that room.": {
			roomID: "room0-id",
			id:     "user0-id",
			events: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr1-id", RoomID: "room1-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr2-id", RoomID: "room2-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr3-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr4-id", RoomID: "room1-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr5-id", RoomID: "room0-id"}},
				}
			},
			expEvents: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr3-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr5-id", RoomID: "room0-id"}},
				}
			},
		},

		"Having a subscription and then unsubscribing on a room, we shouldn't receive events.": {
			roomID:      "room0-id",
			id:          "user0-id",
			unsubscribe: true,
			events: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr1-id", RoomID: "room1-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr2-id", RoomID: "room2-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr3-id", RoomID: "room0-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr4-id", RoomID: "room1-id"}},
					{DiceRoll: model.DiceRoll{ID: "dr5-id", RoomID: "room0-id"}},
				}
			},
			expEvents: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			hub := memory.NewHub(log.Dummy)

			// Subscribe with our check.
			gotEvents := []model.EventDiceRollCreated{}
			err := hub.SubscribeDiceRollCreated(context.TODO(), test.id, test.roomID, func(_ context.Context, e model.EventDiceRollCreated) error {
				gotEvents = append(gotEvents, e)
				return nil
			})
			require.NoError(err)

			// In case we want to unsubscribe after subscription.
			if test.unsubscribe {
				err := hub.UnsubscribeDiceRollCreated(context.TODO(), test.id, test.roomID)
				require.NoError(err)
			}

			// Send
			for _, dr := range test.events() {
				err := hub.NotifyDiceRollCreated(context.TODO(), dr)
				require.NoError(err)
			}

			// TODO(slok): Wait for reception when making the implementation async.

			// Check.
			assert.Equal(test.expEvents(), gotEvents)
		})
	}
}
