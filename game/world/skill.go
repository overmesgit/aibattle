package world

import (
	"errors"
)

func (state *GameState) UseSkill(unit *Unit, target *Position) error {
	if target == nil {
		return errors.New("target is nil")
	}
	skill := unit.Actions.Skill1
	if skill == nil {
		return errors.New("skill is not available")
	}

	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(skill.Range) {
		return errors.New("target is out of range")
	}

	targetUnit, err := state.FindUnit(*target)
	if err != nil {
		return err
	}

	if skill.Effect == HEAL {
		targetUnit.HP += skill.Value
	}
	if skill.Effect == RANGE {
		targetUnit.HP -= skill.Value
	}
	return nil

}
