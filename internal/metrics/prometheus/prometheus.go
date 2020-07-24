package prometheus

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	gohttpmetrics "github.com/slok/go-http-metrics/metrics"
	gohttpmetricsprom "github.com/slok/go-http-metrics/metrics/prometheus"

	"github.com/rollify/rollify/internal/dice"
	"github.com/rollify/rollify/internal/event"
	"github.com/rollify/rollify/internal/http/apiv1"
	"github.com/rollify/rollify/internal/model"
	"github.com/rollify/rollify/internal/room"
	"github.com/rollify/rollify/internal/storage"
	"github.com/rollify/rollify/internal/user"
)

const prefix = "rollify"

// Types used to avoid collisions with the same interface naming.
type httpRecorder gohttpmetrics.Recorder

// Recorder satisfiies multiple interfaces related with metrics measuring
// it will implement Prometheus based metrics backend.
type Recorder struct {
	httpRecorder

	diceRollQuantity                *prometheus.HistogramVec
	dieRollResult                   *prometheus.CounterVec
	diceServiceOPDuration           *prometheus.HistogramVec
	roomServiceOPDuration           *prometheus.HistogramVec
	userServiceOPDuration           *prometheus.HistogramVec
	diceRollRepoOPDuration          *prometheus.HistogramVec
	roomRepoOPDuration              *prometheus.HistogramVec
	userRepoOPDuration              *prometheus.HistogramVec
	notifierOPDuration              *prometheus.HistogramVec
	subscriberSubscribeOPDuration   *prometheus.HistogramVec
	subscriberUnsubscribeOPDuration *prometheus.HistogramVec
	subscriberEvHandleOPDuration    *prometheus.HistogramVec
	subscriberQuantity              *prometheus.GaugeVec
}

// NewRecorder returns a new recorder implementation for prometheus.
func NewRecorder(reg prometheus.Registerer) Recorder {
	r := Recorder{
		httpRecorder: gohttpmetricsprom.NewRecorder(gohttpmetricsprom.Config{Registry: reg}),

		diceRollQuantity: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "dice_roller",
			Name:      "dice_roll_die_quantity",
			Help:      "The quantity of dice on dice rolls.",
			Buckets:   []float64{1, 2, 5, 8, 12, 18, 25, 40, 60, 100},
		}, []string{"roller_type"}),

		dieRollResult: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Subsystem: "dice_roller",
			Name:      "die_roll_results_total",
			Help:      "The total number of die rolls.",
		}, []string{"roller_type", "die_type", "die_side"}),

		diceServiceOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "dice_service",
			Name:      "operation_duration_seconds",
			Help:      "The duration of dice application service.",
		}, []string{"op", "success"}),

		roomServiceOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "room_service",
			Name:      "operation_duration_seconds",
			Help:      "The duration of room application service.",
		}, []string{"op", "success"}),

		userServiceOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "user_service",
			Name:      "operation_duration_seconds",
			Help:      "The duration of user application service.",
		}, []string{"op", "success"}),

		diceRollRepoOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "dice_roll_repository",
			Name:      "operation_duration_seconds",
			Help:      "The duration of dice roll storage repository operations.",
		}, []string{"storage_type", "op", "success"}),

		roomRepoOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "room_repository",
			Name:      "operation_duration_seconds",
			Help:      "The duration of room storage repository operations.",
		}, []string{"storage_type", "op", "success"}),

		userRepoOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "user_repository",
			Name:      "operation_duration_seconds",
			Help:      "The duration of user storage repository operations.",
		}, []string{"storage_type", "op", "success"}),

		notifierOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "notifier",
			Name:      "operation_duration_seconds",
			Help:      "The duration of notifier operations.",
		}, []string{"notifier_type", "op", "success"}),

		subscriberSubscribeOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "subscriber",
			Name:      "subscribe_duration_seconds",
			Help:      "The duration of subscriber subscribe operations.",
		}, []string{"subscriber_type", "subscription", "success"}),

		subscriberUnsubscribeOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "subscriber",
			Name:      "unsubscribe_duration_seconds",
			Help:      "The duration of subscriber unsubscribe operations.",
		}, []string{"subscriber_type", "subscription", "success"}),

		subscriberEvHandleOPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Subsystem: "subscriber",
			Name:      "event_handler_duration_seconds",
			Help:      "The duration of subscriber event handler execution.",
		}, []string{"subscriber_type", "subscription", "success"}),

		subscriberQuantity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prefix,
			Subsystem: "subscriber",
			Name:      "subscribers_subscribed",
			Help:      "The quantity of subscribed subscribers.",
		}, []string{"subscriber_type", "subscription"}),
	}

	reg.MustRegister(
		r.diceRollQuantity,
		r.dieRollResult,
		r.diceServiceOPDuration,
		r.userServiceOPDuration,
		r.roomServiceOPDuration,
		r.diceRollRepoOPDuration,
		r.roomRepoOPDuration,
		r.userRepoOPDuration,
		r.notifierOPDuration,
		r.subscriberSubscribeOPDuration,
		r.subscriberUnsubscribeOPDuration,
		r.subscriberEvHandleOPDuration,
		r.subscriberQuantity,
	)

	return r
}

// MeasureDiceRollQuantity satisfies dice.RollerMetricsRecorder interface.
func (r Recorder) MeasureDiceRollQuantity(ctx context.Context, rollerType string, diceRoll *model.DiceRoll) {
	r.diceRollQuantity.WithLabelValues(rollerType).Observe(float64(len(diceRoll.Dice)))
}

// MeasureDieRollResult satisfies dice.RollerMetricsRecorder interface.
func (r Recorder) MeasureDieRollResult(ctx context.Context, rollerType string, dieRoll *model.DieRoll) {
	r.dieRollResult.WithLabelValues(rollerType, dieRoll.Type.ID(), strconv.Itoa(int(dieRoll.Side))).Inc()
}

// MeasureDiceServiceOpDuration satisfies dice.ServiceMetricsRecorder interface.
func (r Recorder) MeasureDiceServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration) {
	r.diceServiceOPDuration.WithLabelValues(op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureRoomServiceOpDuration satisfies room.ServiceMetricsRecorder interface.
func (r Recorder) MeasureRoomServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration) {
	r.roomServiceOPDuration.WithLabelValues(op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureUserServiceOpDuration satisfies user.ServiceMetricsRecorder interface.
func (r Recorder) MeasureUserServiceOpDuration(ctx context.Context, op string, success bool, t time.Duration) {
	r.userServiceOPDuration.WithLabelValues(op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureDiceRollRepoOpDuration satisfies storage.DiceRollRepositoryMetricsRecorder interface.
func (r Recorder) MeasureDiceRollRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration) {
	r.diceRollRepoOPDuration.WithLabelValues(storageType, op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureRoomRepoOpDuration satisfies storage.RoomRepositoryMetricsRecorder interface.
func (r Recorder) MeasureRoomRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration) {
	r.roomRepoOPDuration.WithLabelValues(storageType, op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureUserRepoOpDuration satisfies storage.UserRepositoryMetricsRecorder interface.
func (r Recorder) MeasureUserRepoOpDuration(ctx context.Context, storageType, op string, success bool, t time.Duration) {
	r.userRepoOPDuration.WithLabelValues(storageType, op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureNotifyOpDuration satisfies event.NotifierMetricsRecorder interface.
func (r Recorder) MeasureNotifyOpDuration(ctx context.Context, notifierType, op string, success bool, t time.Duration) {
	r.notifierOPDuration.WithLabelValues(notifierType, op, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureSubscriberSubscribeOpDuration satisfies event.SubscriberMetricsRecorder interface.
func (r Recorder) MeasureSubscriberSubscribeOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration) {
	r.subscriberSubscribeOPDuration.WithLabelValues(subscriberType, subscription, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureSubscriberUnsubscribeOpDuration satisfies event.SubscriberMetricsRecorder interface.
func (r Recorder) MeasureSubscriberUnsubscribeOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration) {
	r.subscriberUnsubscribeOPDuration.WithLabelValues(subscriberType, subscription, strconv.FormatBool(success)).Observe(t.Seconds())
}

// MeasureSubscriberEventHandleOpDuration satisfies event.SubscriberMetricsRecorder interface.
func (r Recorder) MeasureSubscriberEventHandleOpDuration(ctx context.Context, subscriberType, subscription string, success bool, t time.Duration) {
	r.subscriberEvHandleOPDuration.WithLabelValues(subscriberType, subscription, strconv.FormatBool(success)).Observe(t.Seconds())
}

// AddSubscriberQuantity satisfies event.SubscriberMetricsRecorder interface.
func (r Recorder) AddSubscriberQuantity(ctx context.Context, subscriberType, subscription string, quantity int) {
	r.subscriberQuantity.WithLabelValues(subscriberType, subscription).Add(float64(quantity))
}

var (
	_ apiv1.MetricsRecorder                     = Recorder{}
	_ dice.RollerMetricsRecorder                = Recorder{}
	_ dice.ServiceMetricsRecorder               = Recorder{}
	_ room.ServiceMetricsRecorder               = Recorder{}
	_ user.ServiceMetricsRecorder               = Recorder{}
	_ user.ServiceMetricsRecorder               = Recorder{}
	_ storage.DiceRollRepositoryMetricsRecorder = Recorder{}
	_ storage.RoomRepositoryMetricsRecorder     = Recorder{}
	_ storage.UserRepositoryMetricsRecorder     = Recorder{}
	_ event.NotifierMetricsRecorder             = Recorder{}
	_ event.SubscriberMetricsRecorder           = Recorder{}
)
