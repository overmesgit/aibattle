package world

import (
	"errors"

	"github.com/samber/lo"
)

func (gameState *GameState) UseSkill(unit *Unit, target *Position) ([]int, error) {
	if target == nil {
		return nil, errors.New("target is nil")
	}
	skill := UnitActionMap[unit.Type].Skill1
	if skill == nil {
		return nil, errors.New("skill is not available")
	}

	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(skill.Range) {
		return nil, errors.New("target is out of range")
	}

	targetUnit, err := gameState.FindUnit(*target)
	if err != nil {
		return nil, err
	}

	if skill.Effect == HEAL {
		targetUnit.HP = lo.Min([]int{targetUnit.HP + skill.Value, targetUnit.MaxHP})
	}
	if skill.Effect == RANGE {
		targetUnit.HP -= skill.Value
	}
	return []int{targetUnit.ID}, nil

}
