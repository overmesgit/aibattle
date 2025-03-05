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
	Turn          int                  `json:"turn"`
	Units         []*Unit              `json:"units"`
	Width         int                  `json:"width"`
	Height        int                  `json:"height"`
	UnitActionMap map[string]ActionMap `json:"unit_action_map"`
	IDToUnit      map[int]*Unit
}

func (gameState *GameState) RemoveDeadUnits() {
	gameState.Units = lo.Filter(
		gameState.Units, func(unit *Unit, index int) bool {
			return unit.IsAlive()
		},
	)
}

const (
	Draw int = iota
	TeamA
	TeamB
)

func GetTeamName(teamID int) string {
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
	addUnit(NewWarrior(TeamA, Position{X: 4, Y: 1}))
	addUnit(NewHealer(TeamA, Position{X: 3, Y: 1}))
	addUnit(NewMage(TeamA, Position{X: 2, Y: 1}))
	addUnit(NewRogue(TeamA, Position{X: 1, Y: 1}))

	addUnit(NewWarrior(TeamB, Position{15, 18}))
	addUnit(NewHealer(TeamB, Position{16, 18}))
	addUnit(NewMage(TeamB, Position{17, 18}))
	addUnit(NewRogue(TeamB, Position{X: 18, Y: 18}))

	sort.Slice(
		units, func(i, j int) bool {
			return units[i].Initiative > units[j].Initiative
		},
	)

	unitIDtoUnit := lo.KeyBy(
		units, func(item *Unit) int {
			return item.ID
		},
	)

	return GameState{
		Turn:          0,
		Units:         units,
		Width:         20,
		Height:        20,
		UnitActionMap: UnitActionMap,
		IDToUnit:      unitIDtoUnit,
	}
}

var unitNotFoundErr = errors.New("unit not found")

func (gameState *GameState) FindUnit(position Position) (*Unit, error) {
	for _, unit := range gameState.Units {
		if unit.Position == position && unit.IsAlive() {

			return unit, nil
		}
	}
	return nil, unitNotFoundErr
}

func (gameState *GameState) IsOccupied(position Position) bool {
	_, err := gameState.FindUnit(position)
	if errors.Is(err, unitNotFoundErr) {
		return false
	}
	return true
}

func (gameState *GameState) CopyUnits() []Unit {
	var res []Unit
	for _, unit := range gameState.Units {
		copyUnit := *unit
		res = append(res, copyUnit)
	}
	return res
}

func (gameState *GameState) GetUnitsByIDs(ids []int) []Unit {
	return lo.Map(
		ids, func(id int, index int) Unit {
			unit := gameState.IDToUnit[id]
			if unit == nil {
				return Unit{}
			}
			return *unit
		},
	)
}

func CalculateDistance(pos1, pos2 Position) float64 {
	dx := float64(pos1.X - pos2.X)
	dy := float64(pos1.Y - pos2.Y)
	return math.Sqrt(dx*dx + dy*dy)
}
