package world

import "errors"

func (gameState *GameState) MoveUnit(unit *Unit, target *Position) ([]int, error) {
	if target == nil {
		return nil, errors.New("target is nil")
	}

	// Check boundaries
	if target.X < 0 || target.X >= gameState.Width || target.Y < 0 || target.Y >= gameState.Height {
		return nil, errors.New("target is out of map range")
	}

	// Check distance
	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(UnitActionMap[unit.Type].Move.Distance) {
		return nil, errors.New("target is out of moving range")
	}

	// Check if target position is occupied
	occupied := gameState.IsOccupied(*target)
	if occupied {
		return nil, errors.New("target is occupied")
	}

	unit.Position = *target
	return []int{unit.ID}, nil
}
