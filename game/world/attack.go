package world

import (
	"errors"
	"github.com/samber/lo"
)

func (state *GameState) AttackUnit(unit *Unit, target *Position) error {
	if target == nil {
		return errors.New("target is nil")
	}
	if unit.Actions.Attack1 == nil {
		return errors.New("attack is not available")
	}

	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(unit.Actions.Attack1.Range) {
		return errors.New("target is out of range")
	}

	targetUnit, err := state.FindUnit(*target)
	if err != nil {
		return err
	}

	targetUnit.HP -= unit.Actions.Attack1.Damage
	if targetUnit.HP <= 0 {
		state.Units = lo.Filter(
			state.Units, func(item *Unit, index int) bool {
				return targetUnit.ID != item.ID
			},
		)
	}
	return nil

}
