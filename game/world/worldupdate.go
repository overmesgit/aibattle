package world

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

func (gameState *GameState) UpdateGameState(unit Unit, action UnitAction, prevAction Action) error {
	if action.Action == "" {
		return errors.New("empty action")
	}
	if ifDoubleMove(prevAction, action.Action) {
		return errors.New("same type of actions as first action")
	}
	if !unit.IsAlive() {
		return errors.New(fmt.Sprintf("unit %d is dead", unit.ID))
	}
	switch action.Action {
	case HOLD:
		return nil
	case MOVE:
		err := gameState.MoveUnit(&unit, action.Target)
		if err != nil {
			return err
		}
	case ATTACK1:
		err := gameState.AttackUnit(unit, action.Target)
		if err != nil {
			return err
		}
	case SKILL1:
		err := gameState.UseSkill(unit, action.Target)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unknown action %s", action.Action))
	}
	return nil
}

func ifDoubleMove(prevAction Action, nextAction Action) bool {
	return lo.Count([]Action{prevAction, nextAction}, MOVE) > 1
}
