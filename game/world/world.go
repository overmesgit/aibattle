package world

import (
	"errors"
	"math"
	"sort"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GameState struct {
	Turn   int     `json:"turn"`
	Units  []*Unit `json:"units"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

func GetInitialGameState() GameState {
	units := make([]*Unit, 0)
	addUnit := func(newUnit *Unit) {
		units = append(units, newUnit)
	}

	// Team A starting positions
	addUnit(NewWarrior(TeamA, Position{X: 1, Y: 1}))
	addUnit(NewHealer(TeamA, Position{X: 2, Y: 1}))
	addUnit(NewMage(TeamA, Position{X: 3, Y: 1}))
	addUnit(NewRogue(TeamA, Position{X: 4, Y: 1}))

	addUnit(NewWarrior(TeamB, Position{15, 18}))
	addUnit(NewHealer(TeamB, Position{16, 18}))
	addUnit(NewMage(TeamB, Position{17, 18}))
	addUnit(NewRogue(TeamB, Position{X: 18, Y: 18}))

	sort.Slice(
		units, func(i, j int) bool {
			return units[i].Initiative > units[j].Initiative
		},
	)

	return GameState{
		Turn:   0,
		Units:  units,
		Width:  20,
		Height: 20,
	}
}

func (state *GameState) FindUnit(position Position) (*Unit, error) {
	for _, unit := range state.Units {
		if unit.Position == position {
			return unit, nil
		}
	}
	return nil, errors.New("unit not found")
}

func (state *GameState) CopyUnits() []Unit {
	var res []Unit
	for _, unit := range state.Units {
		copyUnit := *unit
		copyUnit.Actions = ActionMap{}
		res = append(res, copyUnit)
	}
	return res
}

func CalculateDistance(pos1, pos2 Position) float64 {
	dx := float64(pos1.X - pos2.X)
	dy := float64(pos1.Y - pos2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

const (
	TeamA = iota + 1
	TeamB
)

func GetTeamName(teamID int) string {
	switch teamID {
	case TeamA:
		return "TeamA"
	case TeamB:
		return "TeamB"
	default:
		return "TeamX"
	}
}

type Action string

const (
	HOLD    Action = "hold"
	MOVE    Action = "move"
	ATTACK1 Action = "attack1"
	ATTACK2 Action = "attack2"
	SKILL1  Action = "skill1"
	SKILL2  Action = "skill2"
)
