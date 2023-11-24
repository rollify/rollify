package ui

import (
	"context"
	"net/http"
	"strings"

	"github.com/r3labs/sse/v2"
	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/model"
)

const (
	sseStreamPrefixHTML         = "html-"         // Used when we want to use the message as HTML on the DOM (add/replace using HTMX).
	sseStreamPrefixNotification = "notification-" // Used when we only want to be notified without the transportation of all the rendered HTML over the wire..
)

func (u ui) handlerSubscribeDiceRollEvents() http.Handler {
	type subcription struct {
		appSubcriptionCancelFunc func() error
	}

	// TODO(slok): Make it concurrent.
	subcriptionsCancelByRoomID := map[string]subcription{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get room ID from the stream ID.
		roomID := r.URL.Query().Get(queryParamSSEStream)
		roomID = strings.TrimPrefix(roomID, sseStreamPrefixHTML)
		roomID = strings.TrimPrefix(roomID, sseStreamPrefixNotification)

		// If not room subscription already running, create one.
		// TODO(slok): Stop subscription and delete when no connections left.
		_, ok := subcriptionsCancelByRoomID[roomID]
		if ok {
			u.sseServer.ServeHTTP(w, r)
			return
		}

		// Prepare subscriptions.
		subs := subcription{}

		// Create SSE streams.
		// We have 2 different streams for the same event (new dice roll), this is
		// because we want to be notified for the same thing but we will do different things
		// on the client, one will use the data to update the DOM with he received template,
		// and the other one only wants to be notified no matter the event data, if we dont' want
		// the event data and we are receiving lots of HTML over the wire, it's something that
		// impacts performance and transports data without a need.
		// These are the 2 streams:
		//
		// - HTML stream: Will be used to replace directly HTML with the data received from the event
		//	(e.g:  Add directly the HTML of the new dice roll, to the history).
		// - Notifications stream: Will be used as a notification system on the client
		// 	(e.g update a bubble that says the user that there are new rolls).
		u.sseServer.CreateStream(sseStreamPrefixHTML + roomID)
		u.sseServer.CreateStream(sseStreamPrefixNotification + roomID)

		// Start dice rolls subscription.
		modelReq := dice.SubscribeDiceRollCreatedRequest{
			RoomID: roomID,
			EventHandler: func(ctx context.Context, e model.EventDiceRollCreated) error {

				rendered := u.tplRenderS("dice_roll_history_row", u.mapDiceRollToTplModel(e.DiceRoll, true))
				rendered = strings.ReplaceAll(rendered, "\n", "") // https://github.com/r3labs/sse/issues/62.

				// Send to HTML and notification streams.
				u.sseServer.Publish(sseStreamPrefixHTML+roomID, &sse.Event{
					Event: []byte("new_dice_roll"),
					Data:  []byte(rendered),
				})
				u.sseServer.Publish(sseStreamPrefixNotification+roomID, &sse.Event{
					Event: []byte("new_dice_roll"),
					Data:  []byte(e.DiceRoll.ID),
				})

				return nil
			},
		}
		modelResp, err := u.diceAppSvc.SubscribeDiceRollCreated(context.Background(), modelReq)
		if err != nil {
			u.logger.Warningf("Error subscribing SSE to dice roll created events: %s", err)
			return
		}
		subs.appSubcriptionCancelFunc = modelResp.UnsubscribeFunc

		// Store subscriptions data.
		subcriptionsCancelByRoomID[roomID] = subs

		// Continue as always.
		u.sseServer.ServeHTTP(w, r)
	})
}
