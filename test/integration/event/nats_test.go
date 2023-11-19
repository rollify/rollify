//go:build integration

package event_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	eventnats "github.com/rollify/rollify/internal/event/nats"
	"github.com/rollify/rollify/internal/model"
)

const (
	natsAddrEnvVar = "ROLLIFY_INTEGRATION_NATS_ADDR"
)

// TestEventNATS will create a new NATS client and execute all the tests
// related with the NATS Hub implementation.
// Use `ROLLIFY_INTEGRATION_NATS_ADDR` for the NATS address.
func TestEventNATS(t *testing.T) {
	// Create Nats client.
	addr := os.Getenv(natsAddrEnvVar)
	if addr == "" {
		addr = "localhost:4222"
	}

	cli, err := nats.Connect(addr)
	if err != nil {
		t.Fatalf("error connecting nats: %s", err)
	}

	// Execute tests with the NATS client.
	testHubDiceRollCreatedEventsFlow(t, cli)
}

func testHubDiceRollCreatedEventsFlow(t *testing.T, natsCLI *nats.Conn) {
	t0 := time.Now().UTC()

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

		"Having dice rolls on a room we are subscribed, should receive the notifications and correctly mapped.": {
			roomID: "room0-id",
			id:     "user0-id",
			events: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{
						DiceRoll: model.DiceRoll{
							ID:        "dr0-id",
							RoomID:    "room0-id",
							Serial:    42,
							CreatedAt: t0,
							Dice: []model.DieRoll{
								{ID: "dc0", Type: model.DieTypeD4, Side: 2},
								{ID: "dc1", Type: model.DieTypeD6, Side: 4},
							},
						},
					},
				}
			},
			expEvents: func() []model.EventDiceRollCreated {
				return []model.EventDiceRollCreated{
					{
						DiceRoll: model.DiceRoll{
							ID:        "dr0-id",
							RoomID:    "room0-id",
							Serial:    42,
							CreatedAt: t0,
							Dice: []model.DieRoll{
								{ID: "dc0", Type: model.DieTypeD4, Side: 2},
								{ID: "dc1", Type: model.DieTypeD6, Side: 4},
							},
						},
					},
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
					{DiceRoll: model.DiceRoll{ID: "dr0-id", RoomID: "room0-id", Dice: []model.DieRoll{}}},
					{DiceRoll: model.DiceRoll{ID: "dr3-id", RoomID: "room0-id", Dice: []model.DieRoll{}}},
					{DiceRoll: model.DiceRoll{ID: "dr5-id", RoomID: "room0-id", Dice: []model.DieRoll{}}},
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hub, err := eventnats.NewHub(eventnats.HubConfig{
				NATSClient: natsCLI,
				Ctx:        ctx,
			})
			require.NoError(err)

			// Subscribe with our check.
			var gotEventsMu sync.Mutex
			gotEvents := []model.EventDiceRollCreated{}
			err = hub.SubscribeDiceRollCreated(context.TODO(), test.id, test.roomID, func(_ context.Context, e model.EventDiceRollCreated) error {
				gotEventsMu.Lock()
				defer gotEventsMu.Unlock()
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

			// Wait for reception.
			time.Sleep(200 * time.Millisecond)

			// Check.
			gotEventsMu.Lock()
			assert.Equal(test.expEvents(), gotEvents)
			gotEventsMu.Unlock()
		})
	}
}
