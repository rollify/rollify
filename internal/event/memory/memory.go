package memory

import (
	"context"
	"sync"

	"github.com/rollify/rollify/internal/event"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
)

type diceRollCreatedFunc func(context.Context, model.EventDiceRollCreated) error

// Hub implements event.notifier and event.subscriber interfaces with
// a memory implementation. Normally this will be used for single instances
// or as a fake notifier implementation.
//
// TODO(slok): Split in queues and make it async the send process with the reception.
type Hub struct {
	// diceRollCreatedHandlers are the channels stored by roomID, then UserID
	diceRollCreatedHandlers map[string]map[string]diceRollCreatedFunc
	logger                  log.Logger
	mu                      sync.Mutex
}

// NewHub returns a new hub based on a memory implementation.
func NewHub(logger log.Logger) *Hub {
	h := &Hub{
		diceRollCreatedHandlers: map[string]map[string]diceRollCreatedFunc{},
		logger:                  logger.WithKV(log.KV{"service": "memory.Hub"}),
	}

	return h
}

// NotifyDiceRollCreated satisfies event.Notifier interface.
func (h *Hub) NotifyDiceRollCreated(ctx context.Context, e model.EventDiceRollCreated) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	logger := h.logger.WithKV(log.KV{"event": "DiceRollCreated"})

	// Get channel of room.
	handlers, ok := h.diceRollCreatedHandlers[e.DiceRoll.RoomID]
	if !ok {
		handlers = map[string]diceRollCreatedFunc{}
		h.diceRollCreatedHandlers[e.DiceRoll.RoomID] = handlers
	}

	// Broadcast.
	for _, handler := range handlers {
		err := handler(ctx, e)
		if err != nil {
			logger.Errorf("error executing hub event handler : %s", err)
		}
	}

	return nil
}

// SubscribeDiceRollCreated satisfies event.Subscriber interface.
func (h *Hub) SubscribeDiceRollCreated(subscribeID, roomID string, handler func(context.Context, model.EventDiceRollCreated) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	logger := h.logger.WithKV(log.KV{"event": "DiceRollCreated"})

	hs, ok := h.diceRollCreatedHandlers[roomID]
	if !ok {
		hs = map[string]diceRollCreatedFunc{}
	}

	hs[subscribeID] = handler
	h.diceRollCreatedHandlers[roomID] = hs
	logger.Debugf("subscribed to DiceRollCreated events")

	return nil
}

// UnsubscribeDiceRollCreated satisfies event.Subscriber interface.
func (h *Hub) UnsubscribeDiceRollCreated(subscribeID, roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	logger := h.logger.WithKV(log.KV{"event": "DiceRollCreated"})

	hs, ok := h.diceRollCreatedHandlers[roomID]
	if ok {
		delete(hs, subscribeID)
	}

	logger.Debugf("unsubscribed to DiceRollCreated events")
	return nil
}

var (
	_ event.Notifier   = &Hub{}
	_ event.Subscriber = &Hub{}
)
