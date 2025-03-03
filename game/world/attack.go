package world

import (
	"errors"
)

func (gameState *GameState) AttackUnit(unit *Unit, target *Position) ([]int, error) {
	if target == nil {
		return nil, errors.New("target is nil")
	}
	attack := UnitActionMap[unit.Type].Attack1
	if attack == nil {
		return nil, errors.New("attack is not available")
	}

	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(attack.Range) {
		return nil, errors.New("target is out of range")
	}

	targetUnit, err := gameState.FindUnit(*target)
	if err != nil {
		return nil, err
	}

	targetUnit.HP -= attack.Damage
	return []int{targetUnit.ID}, nil

}
