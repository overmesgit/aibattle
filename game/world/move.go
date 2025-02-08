package world

import "errors"

func (state *GameState) MoveUnit(unit *Unit, target *Position) error {
	if target == nil {
		return errors.New("target is nil")
	}

	// Check boundaries
	if target.X < 0 || target.X >= state.Width || target.Y < 0 || target.Y >= state.Height {
		return errors.New("target is out of map range")
	}

	// Check distance
	distance := CalculateDistance(unit.Position, *target)
	if distance > float64(unit.Actions.Move.Distance) {
		return errors.New("target is out of moving range")
	}

	// Check if target position is occupied
	occupied := state.IsOccupied(*target)
	if occupied {
		return errors.New("target is occupied")
	}

	unit.Position = *target
	return nil
}
