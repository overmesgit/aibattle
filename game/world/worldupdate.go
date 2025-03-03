package world

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

func (gameState *GameState) UpdateGameState(
	unit *Unit, action UnitAction, prevAction Action,
) ([]int, error) {
	if action.Action == "" {
		return nil, errors.New("empty action")
	}
	if ifDoubleMove(prevAction, action.Action) {
		return nil, errors.New("same type of actions as first action")
	}
	if !unit.IsAlive() {
		return nil, errors.New(fmt.Sprintf("unit %d is dead", unit.ID))
	}
	switch action.Action {
	case HOLD:
		return nil, nil
	case MOVE:
		return gameState.MoveUnit(unit, action.Target)
	case ATTACK1:
		return gameState.AttackUnit(unit, action.Target)
	case SKILL1:
		return gameState.UseSkill(unit, action.Target)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown action %s", action.Action))
	}
}

func ifDoubleMove(prevAction Action, nextAction Action) bool {
	return lo.Count([]Action{prevAction, nextAction}, MOVE) > 1
}
