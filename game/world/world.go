package world

import (
	"errors"
	"math"
	"sort"

	"github.com/samber/lo"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GameState struct {
	Turn          int                `json:"turn"`
	Units         []*Unit            `json:"units"`
	Width         int                `json:"width"`
	Height        int                `json:"height"`
	UnitActionMap map[Type]ActionMap `json:"unit_action_map"`
}

func (state GameState) RemoveDeadUnits() {
	state.Units = lo.Filter(
		state.Units, func(unit *Unit, index int) bool {
			return unit.IsAlive()
		},
	)
}

type Team int

const (
	Draw Team = iota
	TeamA
	TeamB
)

func GetTeamName(teamID Team) string {
	switch teamID {
	case Draw:
		return "Draw"
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

	addUnit(NewWarrior(TeamB, Position{18, 18}))
	addUnit(NewHealer(TeamB, Position{17, 18}))
	addUnit(NewMage(TeamB, Position{16, 18}))
	addUnit(NewRogue(TeamB, Position{X: 15, Y: 18}))

	sort.Slice(
		units, func(i, j int) bool {
			return units[i].Initiative > units[j].Initiative
		},
	)

	return GameState{
		Turn:          0,
		Units:         units,
		Width:         20,
		Height:        20,
		UnitActionMap: UnitActionMap,
	}
}

var unitNotFoundErr = errors.New("unit not found")

func (state *GameState) FindUnit(position Position) (*Unit, error) {
	for _, unit := range state.Units {
		if unit.Position == position && unit.IsAlive() {

			return unit, nil
		}
	}
	return nil, unitNotFoundErr
}

func (state *GameState) IsOccupied(position Position) bool {
	_, err := state.FindUnit(position)
	if errors.Is(err, unitNotFoundErr) {
		return false
	}
	return true
}

func (state *GameState) CopyUnits() []Unit {
	var res []Unit
	for _, unit := range state.Units {
		copyUnit := *unit
		res = append(res, copyUnit)
	}
	return res
}

func CalculateDistance(pos1, pos2 Position) float64 {
	dx := float64(pos1.X - pos2.X)
	dy := float64(pos1.Y - pos2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}
