package memory

import (
	"context"
	"sync"

	"github.com/rollify/rollify/internal/event"
	"github.com/rollify/rollify/internal/model"
)

type diceRollCreatedFunc func(context.Context, model.DiceRoll) error

// Hub implements event.notifier and event.subscriber interfaces with
// a memory implementation. Normally this will be used for single instances
// or as a fake notifier implementation.
//
// TODO(slok): Split in queues and make it async the send process with the reception.
type Hub struct {
	// diceRollCreatedHandlers are the channels stored by roomID, then UserID
	diceRollCreatedHandlers map[string]map[string]diceRollCreatedFunc
	mu                      sync.Mutex
}

// NewHub returns a new hub based on a memory implementation.
func NewHub() *Hub {
	h := &Hub{
		diceRollCreatedHandlers: map[string]map[string]diceRollCreatedFunc{},
	}

	return h
}

// NotifyDiceRollCreated satisfies event.Notifier interface.
func (h *Hub) NotifyDiceRollCreated(ctx context.Context, d model.DiceRoll) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get channel of room.
	handlers, ok := h.diceRollCreatedHandlers[d.RoomID]
	if !ok {
		handlers = map[string]diceRollCreatedFunc{}
		h.diceRollCreatedHandlers[d.RoomID] = handlers
	}

	// Broadcast.
	for _, h := range handlers {
		h(ctx, d)
	}

	return nil
}

// SubscribeDiceRollCreated satisfies event.Subscriber interface.
func (h *Hub) SubscribeDiceRollCreated(roomID, userID string, handler func(context.Context, model.DiceRoll) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	hs, ok := h.diceRollCreatedHandlers[roomID]
	if !ok {
		hs = map[string]diceRollCreatedFunc{}
	}

	hs[userID] = handler
	h.diceRollCreatedHandlers[roomID] = hs

	return nil
}

// UnsubscribeDiceRollCreated satisfies event.Subscriber interface.
func (h *Hub) UnsubscribeDiceRollCreated(roomID, userID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	hs, ok := h.diceRollCreatedHandlers[roomID]
	if ok {
		delete(hs, userID)
	}
	return nil
}

var (
	_ event.Notifier   = &Hub{}
	_ event.Subscriber = &Hub{}
)
