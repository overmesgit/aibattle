package world

import (
	"errors"
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
	return nil

}
