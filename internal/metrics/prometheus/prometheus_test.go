package prometheus_test

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metrics "github.com/rollify/rollify/internal/metrics/prometheus"
	"github.com/rollify/rollify/internal/model"
)

func TestRecorder(t *testing.T) {
	tests := map[string]struct {
		measure    func(r metrics.Recorder)
		expMetrics []string
	}{
		"Measure dice roller dice roll dice quantity.": {
			measure: func(r metrics.Recorder) {
				r.MeasureDiceRollQuantity(context.TODO(), "test-1", &model.DiceRoll{})
				r.MeasureDiceRollQuantity(context.TODO(), "test-2", &model.DiceRoll{
					Dice: []model.DieRoll{{}, {}, {}},
				})
				r.MeasureDiceRollQuantity(context.TODO(), "test-1", &model.DiceRoll{
					Dice: []model.DieRoll{{}, {}},
				})
				r.MeasureDiceRollQuantity(context.TODO(), "test-1", &model.DiceRoll{
					Dice: []model.DieRoll{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
				})
			},
			expMetrics: []string{
				`# HELP rollify_dice_roller_dice_roll_die_quantity The quantity of dice on dice rolls.`,
				`# TYPE rollify_dice_roller_dice_roll_die_quantity histogram`,

				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="1"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="2"} 2`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="5"} 2`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="8"} 2`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="12"} 2`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="18"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="25"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="40"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="60"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="100"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-1",le="+Inf"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_sum{roller_type="test-1"} 19`,
				`rollify_dice_roller_dice_roll_die_quantity_count{roller_type="test-1"} 3`,

				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="1"} 0`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="2"} 0`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="5"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="8"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="12"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="18"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="25"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="40"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="60"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="100"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_bucket{roller_type="test-2",le="+Inf"} 1`,
				`rollify_dice_roller_dice_roll_die_quantity_sum{roller_type="test-2"} 3`,
				`rollify_dice_roller_dice_roll_die_quantity_count{roller_type="test-2"} 1`,
			},
		},

		"Measure dice roller die roll result.": {
			measure: func(r metrics.Recorder) {
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD4, Side: 2})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD6, Side: 2})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD6, Side: 4})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD6, Side: 4})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD6, Side: 4})
				r.MeasureDieRollResult(context.TODO(), "test-2", &model.DieRoll{Type: model.DieTypeD6, Side: 5})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD8, Side: 7})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD10, Side: 7})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD12, Side: 10})
				r.MeasureDieRollResult(context.TODO(), "test-1", &model.DieRoll{Type: model.DieTypeD20, Side: 8})
				r.MeasureDieRollResult(context.TODO(), "test-2", &model.DieRoll{Type: model.DieTypeD20, Side: 17})
				r.MeasureDieRollResult(context.TODO(), "test-2", &model.DieRoll{Type: model.DieTypeD20, Side: 17})

			},
			expMetrics: []string{
				`# HELP rollify_dice_roller_die_roll_results_total The total number of die rolls.`,
				`# TYPE rollify_dice_roller_die_roll_results_total counter`,
				`rollify_dice_roller_die_roll_results_total{die_side="10",die_type="d12",roller_type="test-1"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="17",die_type="d20",roller_type="test-2"} 2`,
				`rollify_dice_roller_die_roll_results_total{die_side="2",die_type="d4",roller_type="test-1"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="2",die_type="d6",roller_type="test-1"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="4",die_type="d6",roller_type="test-1"} 3`,
				`rollify_dice_roller_die_roll_results_total{die_side="5",die_type="d6",roller_type="test-2"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="7",die_type="d10",roller_type="test-1"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="7",die_type="d8",roller_type="test-1"} 1`,
				`rollify_dice_roller_die_roll_results_total{die_side="8",die_type="d20",roller_type="test-1"} 1`,
			},
		},

		"Measure dice app service operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureDiceServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureDiceServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureDiceServiceOpDuration(context.TODO(), "op1", false, 267*time.Millisecond)
				r.MeasureDiceServiceOpDuration(context.TODO(), "op1", true, 6*time.Second)
				r.MeasureDiceServiceOpDuration(context.TODO(), "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_dice_service_operation_duration_seconds The duration of dice application service.`,
				`# TYPE rollify_dice_service_operation_duration_seconds histogram`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.005"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.01"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.025"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.05"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.1"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.25"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="0.5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="1"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="2.5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="10"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="false",le="+Inf"} 1`,
				`rollify_dice_service_operation_duration_seconds_count{op="op1",success="false"} 1`,

				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.005"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.01"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.025"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.05"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.1"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.25"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.5"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="1"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="2.5"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="5"} 2`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="10"} 3`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op1",success="true",le="+Inf"} 3`,
				`rollify_dice_service_operation_duration_seconds_count{op="op1",success="true"} 3`,

				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.005"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.01"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.025"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.05"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.1"} 0`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.25"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="1"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="2.5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="5"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="10"} 1`,
				`rollify_dice_service_operation_duration_seconds_bucket{op="op2",success="false",le="+Inf"} 1`,
				`rollify_dice_service_operation_duration_seconds_count{op="op2",success="false"} 1`,
			},
		},

		"Measure room app service operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureRoomServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureRoomServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureRoomServiceOpDuration(context.TODO(), "op1", true, 6*time.Second)
				r.MeasureRoomServiceOpDuration(context.TODO(), "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_room_service_operation_duration_seconds The duration of room application service.`,
				`# TYPE rollify_room_service_operation_duration_seconds histogram`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.005"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.01"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.025"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.05"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.1"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.25"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.5"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="1"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="2.5"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="5"} 2`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="10"} 3`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op1",success="true",le="+Inf"} 3`,
				`rollify_room_service_operation_duration_seconds_count{op="op1",success="true"} 3`,

				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.005"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.01"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.025"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.05"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.1"} 0`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.25"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.5"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="1"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="2.5"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="5"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="10"} 1`,
				`rollify_room_service_operation_duration_seconds_bucket{op="op2",success="false",le="+Inf"} 1`,
				`rollify_room_service_operation_duration_seconds_count{op="op2",success="false"} 1`,
			},
		},

		"Measure user app service operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureUserServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureUserServiceOpDuration(context.TODO(), "op1", true, 55*time.Millisecond)
				r.MeasureUserServiceOpDuration(context.TODO(), "op1", true, 6*time.Second)
				r.MeasureUserServiceOpDuration(context.TODO(), "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_user_service_operation_duration_seconds The duration of user application service.`,
				`# TYPE rollify_user_service_operation_duration_seconds histogram`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.005"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.01"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.025"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.05"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.1"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.25"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="0.5"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="1"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="2.5"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="5"} 2`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="10"} 3`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op1",success="true",le="+Inf"} 3`,
				`rollify_user_service_operation_duration_seconds_count{op="op1",success="true"} 3`,

				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.005"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.01"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.025"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.05"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.1"} 0`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.25"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="0.5"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="1"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="2.5"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="5"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="10"} 1`,
				`rollify_user_service_operation_duration_seconds_bucket{op="op2",success="false",le="+Inf"} 1`,
				`rollify_user_service_operation_duration_seconds_count{op="op2",success="false"} 1`,
			},
		},

		"Measure dice roll repo operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureDiceRollRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureDiceRollRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureDiceRollRepoOpDuration(context.TODO(), "t1", "op1", true, 6*time.Second)
				r.MeasureDiceRollRepoOpDuration(context.TODO(), "t2", "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_dice_roll_repository_operation_duration_seconds The duration of dice roll storage repository operations.`,
				`# TYPE rollify_dice_roll_repository_operation_duration_seconds histogram`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.005"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.01"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.025"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.05"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.1"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.25"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.5"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="1"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="2.5"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="5"} 2`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="10"} 3`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="+Inf"} 3`,
				`rollify_dice_roll_repository_operation_duration_seconds_count{op="op1",storage_type="t1",success="true"} 3`,

				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.005"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.01"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.025"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.05"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.1"} 0`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.25"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.5"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="1"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="2.5"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="5"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="10"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="+Inf"} 1`,
				`rollify_dice_roll_repository_operation_duration_seconds_count{op="op2",storage_type="t2",success="false"} 1`,
			},
		},

		"Measure room repo operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureRoomRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureRoomRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureRoomRepoOpDuration(context.TODO(), "t1", "op1", true, 6*time.Second)
				r.MeasureRoomRepoOpDuration(context.TODO(), "t2", "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_room_repository_operation_duration_seconds The duration of room storage repository operations.`,
				`# TYPE rollify_room_repository_operation_duration_seconds histogram`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.005"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.01"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.025"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.05"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.1"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.25"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.5"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="1"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="2.5"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="5"} 2`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="10"} 3`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="+Inf"} 3`,
				`rollify_room_repository_operation_duration_seconds_count{op="op1",storage_type="t1",success="true"} 3`,

				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.005"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.01"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.025"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.05"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.1"} 0`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.25"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.5"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="1"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="2.5"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="5"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="10"} 1`,
				`rollify_room_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="+Inf"} 1`,
				`rollify_room_repository_operation_duration_seconds_count{op="op2",storage_type="t2",success="false"} 1`,
			},
		},

		"Measure user repo operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureUserRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureUserRepoOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureUserRepoOpDuration(context.TODO(), "t1", "op1", true, 6*time.Second)
				r.MeasureUserRepoOpDuration(context.TODO(), "t2", "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_user_repository_operation_duration_seconds The duration of user storage repository operations.`,
				`# TYPE rollify_user_repository_operation_duration_seconds histogram`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.005"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.01"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.025"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.05"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.1"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.25"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="0.5"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="1"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="2.5"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="5"} 2`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="10"} 3`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op1",storage_type="t1",success="true",le="+Inf"} 3`,
				`rollify_user_repository_operation_duration_seconds_count{op="op1",storage_type="t1",success="true"} 3`,

				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.005"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.01"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.025"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.05"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.1"} 0`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.25"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="0.5"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="1"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="2.5"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="5"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="10"} 1`,
				`rollify_user_repository_operation_duration_seconds_bucket{op="op2",storage_type="t2",success="false",le="+Inf"} 1`,
				`rollify_user_repository_operation_duration_seconds_count{op="op2",storage_type="t2",success="false"} 1`,
			},
		},

		"Measure notifier operation duration.": {
			measure: func(r metrics.Recorder) {
				r.MeasureNotifyOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureNotifyOpDuration(context.TODO(), "t1", "op1", true, 55*time.Millisecond)
				r.MeasureNotifyOpDuration(context.TODO(), "t1", "op1", true, 6*time.Second)
				r.MeasureNotifyOpDuration(context.TODO(), "t2", "op2", false, 143*time.Millisecond)
			},
			expMetrics: []string{
				`# HELP rollify_notifier_operation_duration_seconds The duration of notifier operations.`,
				`# TYPE rollify_notifier_operation_duration_seconds histogram`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.005"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.01"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.025"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.05"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.1"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.25"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="0.5"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="1"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="2.5"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="5"} 2`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="10"} 3`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t1",op="op1",success="true",le="+Inf"} 3`,
				`rollify_notifier_operation_duration_seconds_count{notifier_type="t1",op="op1",success="true"} 3`,

				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.005"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.01"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.025"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.05"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.1"} 0`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.25"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="0.5"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="1"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="2.5"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="5"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="10"} 1`,
				`rollify_notifier_operation_duration_seconds_bucket{notifier_type="t2",op="op2",success="false",le="+Inf"} 1`,
				`rollify_notifier_operation_duration_seconds_count{notifier_type="t2",op="op2",success="false"} 1`,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			reg := prometheus.NewRegistry()
			rec := metrics.NewRecorder(reg)

			test.measure(rec)

			h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
			allMetrics, err := ioutil.ReadAll(w.Result().Body)
			require.NoError(err)

			// Check metrics.
			for _, expMetric := range test.expMetrics {
				assert.Contains(string(allMetrics), expMetric)
			}
		})
	}
}
