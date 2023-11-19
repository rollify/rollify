package nats

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/rollify/rollify/internal/event"
	"github.com/rollify/rollify/internal/log"
	"github.com/rollify/rollify/internal/model"
)

const (
	natsSubjectDiceRollCreated = "rollify.room.diceroll.create"
)

// Client is the client used for NATS connections.
type Client interface {
	ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error)
	Publish(subj string, data []byte) error
}

type diceRollCreatedFunc = func(context.Context, model.EventDiceRollCreated) error

// HubConfig is the hub configuration.
type HubConfig struct {
	// NATSClient is the client to connect to NATS server.
	NATSClient Client
	// Ctx is used to control the event loop, when the context is done,
	// the Hub will stop subscriptions and event handling and a new Hub is required.
	Ctx context.Context
	// Logger is the logger.
	Logger log.Logger
}

func (c *HubConfig) defaults() error {
	if c.NATSClient == nil {
		return fmt.Errorf("the NATS client is required")
	}

	if c.Ctx == nil {
		return fmt.Errorf("context is required")
	}

	if c.Logger == nil {
		c.Logger = log.Dummy
	}

	c.Logger = c.Logger.WithKV(log.KV{"service": "nats.Hub"})

	return nil
}

// Hub implements event.notifier and event.subscriber interfaces with
// a NATS pubsub implementation.
// The Hub needs to be started using Run.
type Hub struct {
	cli    Client
	logger log.Logger

	diceRollCreatedHandlers map[string]map[string]diceRollCreatedFunc
	diceRollCreatedChan     chan *nats.Msg
	diceRollCreatedSubs     *nats.Subscription
	mu                      sync.Mutex
}

// NewHub returns a new hub based on a memory implementation.
func NewHub(cfg HubConfig) (*Hub, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	h := &Hub{
		cli:    cfg.NATSClient,
		logger: cfg.Logger,

		diceRollCreatedHandlers: map[string]map[string]diceRollCreatedFunc{},
		diceRollCreatedChan:     make(chan *nats.Msg, 15),
	}

	// Subscribe and run event handling.
	err = h.subscribeOnNATSSubects()
	if err != nil {
		return nil, fmt.Errorf("could not subscribe on NATS server: %w", err)
	}
	go h.run(cfg.Ctx)

	return h, nil
}

// run will subscribe to NATS server and start the event handling loop.
// Is a blocking operation.
// The function will end when the context is done.
func (h *Hub) run(ctx context.Context) {
	// Unsubscribe when finished.
	defer func() {
		err := h.unsubscribeOnNATSSubects()
		if err != nil {
			h.logger.Warningf("could not unsubscribe on NATS server: %w", err)
		}
	}()

	// Start event handling loop.
	h.logger.Infof("started NATS event receiver loop")
	defer h.logger.Infof("stopped NATS event receiver loop")

	for {
		loopCtx := context.Background()
		select {
		case <-ctx.Done():
			h.logger.Infof("stopping NATS hub by context done")

		case msg := <-h.diceRollCreatedChan:
			h.logger.Debugf("diceRollCreated NATS event received, broadcasting")
			err := h.handleDiceRollCreatedEvent(loopCtx, msg.Data)
			if err != nil {
				h.logger.Errorf("could not handle diceRollCreated event: %s", err)
			}
		}
	}
}

func (h *Hub) subscribeOnNATSSubects() error {
	// Subscribe channels.
	sub, err := h.cli.ChanSubscribe(natsSubjectDiceRollCreated, h.diceRollCreatedChan)
	if err != nil {
		return fmt.Errorf("could not subscribe on created dice roll event subject: %w", err)
	}
	h.diceRollCreatedSubs = sub

	return nil
}

func (h *Hub) unsubscribeOnNATSSubects() error {
	err := h.diceRollCreatedSubs.Unsubscribe()
	if err != nil {
		return fmt.Errorf("could not unsubscribe on created dice roll event subject: %w", err)
	}

	return nil
}

// NotifyDiceRollCreated satisfies event.Notifier interface by pusblishing the event
// in a NATS pubsub stream, serialized in JSON.
func (h *Hub) NotifyDiceRollCreated(ctx context.Context, e model.EventDiceRollCreated) error {
	bs, err := mapModelToBytesEventDiceRollCreated(e)
	if err != nil {
		return fmt.Errorf("could not marshall event: %w", err)
	}

	h.logger.Debugf("diceRollCreated NATS event published")
	err = h.cli.Publish(natsSubjectDiceRollCreated, bs)
	if err != nil {
		return fmt.Errorf("could not pusblish message on NATS: %w", err)
	}

	return nil
}

func (h *Hub) handleDiceRollCreatedEvent(ctx context.Context, data []byte) error {
	e, err := mapBytesToModelEventDiceRollCreated(data)
	if err != nil {
		return fmt.Errorf("could not unmarshall event: %w", err)
	}

	logger := h.logger.WithKV(log.KV{"event": "DiceRollCreated"})

	// Get subscribed handlers.
	h.mu.Lock()
	handlers, ok := h.diceRollCreatedHandlers[e.DiceRoll.RoomID]
	h.mu.Unlock()
	if !ok {
		handlers = map[string]diceRollCreatedFunc{}
		h.diceRollCreatedHandlers[e.DiceRoll.RoomID] = handlers
	}

	// Broadcast to al subscribers.
	for _, handler := range handlers {
		err := handler(ctx, *e)
		if err != nil {
			logger.Errorf("error executing hub event handler : %s", err)
		}
	}

	return nil
}

// SubscribeDiceRollCreated satisfies event.Subscriber interface.
func (h *Hub) SubscribeDiceRollCreated(ctx context.Context, subscribeID, roomID string, handler func(context.Context, model.EventDiceRollCreated) error) error {
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
func (h *Hub) UnsubscribeDiceRollCreated(ctx context.Context, subscribeID, roomID string) error {
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
