package world

import (
	"sync"
)

const (
	WARRIOR = "warrior"
	HEALER  = "healer"
	MAGE    = "mage"
	ROGUE   = "rogue"
)

type Attack struct {
	Range  int `json:"range,omitempty"`
	Damage int `json:"damage,omitempty"`
}

type Move struct {
	Distance int `json:"distance,omitempty"`
}

const (
	HEAL  = "heal"
	RANGE = "range"
)

type Skill struct {
	Effect string `json:"effect"`
	Range  int    `json:"range"`
	Value  int    `json:"value"`
	Name   string `json:"name"`
}

type ActionMap struct {
	Move    *Move   `json:"move,omitempty"`
	Hold    *Move   `json:"hold,omitempty"`
	Attack1 *Attack `json:"attack1,omitempty"`
	Skill1  *Skill  `json:"skill1,omitempty"`
	Skill2  *Skill  `json:"skill2,omitempty"`
	Skill3  *Skill  `json:"skill3,omitempty"`
}

type Unit struct {
	ID         int      `json:"id"`
	Team       int      `json:"team"`
	Type       string   `json:"type"`
	Initiative int      `json:"initiative"`
	HP         int      `json:"hp"`
	MaxHP      int      `json:"maxHp"`
	Position   Position `json:"position"`
}

var UnitActionMap = map[string]ActionMap{
	WARRIOR: {
		Move:    &Move{3},
		Hold:    &Move{},
		Attack1: &Attack{1, 30},
	},
	HEALER: {
		Move:    &Move{2},
		Hold:    &Move{},
		Attack1: &Attack{1, 10},
		Skill1:  &Skill{HEAL, 5, 30, "heal"},
	},
	MAGE: {
		Move:    &Move{2},
		Hold:    &Move{},
		Attack1: &Attack{1, 10},
		Skill1:  &Skill{RANGE, 4, 40, "firebolt"},
	},
	ROGUE: {
		Move:    &Move{4},
		Hold:    &Move{},
		Attack1: &Attack{1, 25},
	},
}

func (u *Unit) IsAlive() bool {
	return u.HP > 0
}

var counter = sync.OnceValue(NewCounter)()

func NewWarrior(team int, position Position) *Unit {
	return &Unit{
		ID:         counter.Get(),
		Team:       team,
		Type:       WARRIOR,
		Initiative: 1,
		HP:         200,
		MaxHP:      200,
		Position:   position,
	}
}

func NewHealer(team int, position Position) *Unit {
	return &Unit{
		ID:         counter.Get(),
		Team:       team,
		Type:       HEALER,
		Initiative: 2,
		HP:         100,
		MaxHP:      100,
		Position:   position,
	}
}

func NewMage(team int, position Position) *Unit {
	return &Unit{
		ID:         counter.Get(),
		Team:       team,
		Type:       MAGE,
		Initiative: 3,
		HP:         120,
		MaxHP:      120,
		Position:   position,
	}
}

func NewRogue(team int, position Position) *Unit {
	return &Unit{
		ID:         counter.Get(),
		Team:       team,
		Type:       ROGUE,
		Initiative: 4,
		HP:         130,
		MaxHP:      130,
		Position:   position,
	}
}

type Counter struct {
	i int
}

func (c *Counter) Get() int {
	c.i++
	return c.i
}

func NewCounter() *Counter {
	return &Counter{}
}
