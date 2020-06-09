package model

// DieType is the physical type of Die structure.
type DieType interface {
	// ID returns the id of the Die type.
	ID() string
	// Sides returns the number of sides our Die has.
	Sides() uint
}

type d4 int

func (d d4) ID() string  { return "d4" }
func (d d4) Sides() uint { return 4 }

type d6 int

func (d d6) ID() string  { return "d6" }
func (d d6) Sides() uint { return 6 }

type d8 int

func (d d8) ID() string  { return "d8" }
func (d d8) Sides() uint { return 8 }

type d10 int

func (d d10) ID() string  { return "d10" }
func (d d10) Sides() uint { return 10 }

type d12 int

func (d d12) ID() string  { return "d12" }
func (d d12) Sides() uint { return 12 }

type d20 int

func (d d20) ID() string  { return "d20" }
func (d d20) Sides() uint { return 20 }

// Die types.
const (
	DieTypeD4  = d4(0)
	DieTypeD6  = d6(0)
	DieTypeD8  = d8(0)
	DieTypeD10 = d10(0)
	DieTypeD12 = d12(0)
	DieTypeD20 = d20(0)
)
