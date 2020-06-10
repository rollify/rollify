package dice

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/rollify/rollify/internal/model"
)

// Roller knows how to  process dice rolls.
type Roller interface {
	Roll(ctx context.Context, d *model.DiceRoll) error
}

//go:generate mockery -case underscore -output dicemock -outpkg dicemock -name Roller

// RollerFunc is a helper type to create rollers from funcs.
type RollerFunc func(ctx context.Context, d *model.DiceRoll) error

// Roll satisfies Roller interface.
func (f RollerFunc) Roll(ctx context.Context, d *model.DiceRoll) error { return f(ctx, d) }

var _ Roller = RollerFunc(nil)

type randomRoller struct {
	mu   sync.Mutex
	rand *rand.Rand
}

// NewRandomRoller returns a roller that knows how to roll dice using a random implementation
// setting the side that we got from the roll.
func NewRandomRoller() Roller {
	// Crea a different seed each time.
	src := rand.NewSource(time.Now().UTC().UnixNano())
	return &randomRoller{
		rand: rand.New(src),
	}
}

func (r *randomRoller) Roll(ctx context.Context, dr *model.DiceRoll) error {
	// Roll each of the dice using our seeded rand.
	ds := make([]model.DieRoll, 0, len(dr.Dice))
	for _, d := range dr.Dice {
		d.Side = uint(r.randInt(int(d.Type.Sides())))
		ds = append(ds, d)
	}

	dr.Dice = ds

	return nil
}

func (r *randomRoller) randInt(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(n)
}
