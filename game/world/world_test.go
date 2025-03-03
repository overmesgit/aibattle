package world

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameStateBasics(t *testing.T) {
	// Test initial game state
	state := GetInitialGameState()
	expectedState := GameState{
		Turn:          0,
		Width:         20,
		Height:        20,
		Units:         state.Units,
		UnitActionMap: state.UnitActionMap,
		IDToUnit:      state.IDToUnit,
	}
	assert.Equal(t, expectedState, state)
}

func TestUnitCreation(t *testing.T) {
	// Test warrior creation
	position := Position{X: 1, Y: 1}
	warrior := NewWarrior(TeamA, position)
	expectedWarrior := &Unit{
		ID:         warrior.ID, // Can't predict exact ID
		Team:       TeamA,
		Type:       WARRIOR,
		Initiative: 1,
		HP:         200,
		MaxHP:      200,
		Position:   position,
	}
	assert.Equal(t, expectedWarrior, warrior)
	assert.True(t, warrior.IsAlive())

	// Test healer creation
	position = Position{X: 5, Y: 5}
	healer := NewHealer(TeamB, position)
	expectedHealer := &Unit{
		ID:         healer.ID, // Can't predict exact ID
		Team:       TeamB,
		Type:       HEALER,
		Initiative: 2,
		HP:         100,
		MaxHP:      100,
		Position:   position,
	}
	assert.Equal(t, expectedHealer, healer)
	assert.True(t, healer.IsAlive())
}

func TestMoveUnit(t *testing.T) {
	state := GetInitialGameState()
	unit := state.Units[0]
	originalPos := unit.Position

	// Test valid move
	target := &Position{X: originalPos.X + 1, Y: originalPos.Y + 1}
	affectedUnits, err := state.MoveUnit(unit, target)
	assert.NoError(t, err)
	assert.Equal(t, *target, unit.Position)
	assert.Equal(t, []int{unit.ID}, affectedUnits)

	// Test move out of range
	target = &Position{X: originalPos.X + 10, Y: originalPos.Y + 10}
	affectedUnits, err = state.MoveUnit(unit, target)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of moving range")
	assert.Nil(t, affectedUnits)

	// Test move out of map
	target = &Position{X: -1, Y: -1}
	affectedUnits, err = state.MoveUnit(unit, target)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of map range")
	assert.Nil(t, affectedUnits)

	// Test move to occupied position
	occupiedPos := &Position{X: state.Units[2].Position.X, Y: state.Units[2].Position.Y}
	affectedUnits, err = state.MoveUnit(unit, occupiedPos)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target is occupied")
	assert.Nil(t, affectedUnits)
}

func TestAttackUnit(t *testing.T) {
	state := GetInitialGameState()

	// Set up attacker and target
	attacker := state.Units[0]   // TeamA
	targetUnit := state.Units[4] // TeamB

	// Move target within range
	attacker.Position = Position{X: 2, Y: 2}
	targetUnit.Position = Position{X: 3, Y: 2}

	// Get initial HP
	initialHP := targetUnit.HP

	// Test valid attack
	affectedUnits, err := state.AttackUnit(attacker, &targetUnit.Position)
	assert.NoError(t, err)
	assert.Less(t, targetUnit.HP, initialHP)
	assert.Equal(t, []int{targetUnit.ID}, affectedUnits)

	// Test attack out of range
	attacker.Position = Position{X: 1, Y: 1}
	targetUnit.Position = Position{X: 10, Y: 10}
	affectedUnits, err = state.AttackUnit(attacker, &targetUnit.Position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
	assert.Nil(t, affectedUnits)

	// Test attack non-existent unit
	emptyPos := &Position{X: 15, Y: 15}
	affectedUnits, err = state.AttackUnit(attacker, emptyPos)
	assert.Error(t, err)
	assert.Nil(t, affectedUnits)
}

func TestUseSkill(t *testing.T) {
	state := GetInitialGameState()

	// Find healer
	var healer *Unit
	var teammate *Unit

	for _, unit := range state.Units {
		if unit.Type == HEALER && unit.Team == TeamA {
			healer = unit
		} else if unit.Type == WARRIOR && unit.Team == TeamA {
			teammate = unit
		}
	}

	if healer == nil || teammate == nil {
		t.Fatal("Could not find healer and teammate")
	}

	// Damage teammate
	teammate.HP = 50

	// Position them close together
	healer.Position = Position{X: 5, Y: 5}
	teammate.Position = Position{X: 6, Y: 5}

	// Test healing skill
	initialHP := teammate.HP
	affectedUnits, err := state.UseSkill(healer, &teammate.Position)
	assert.NoError(t, err)
	assert.Greater(t, teammate.HP, initialHP)
	assert.Equal(t, []int{teammate.ID}, affectedUnits)

	// Find mage
	var mage *Unit
	var enemy *Unit

	for _, unit := range state.Units {
		if unit.Type == MAGE && unit.Team == TeamA {
			mage = unit
		} else if unit.Team == TeamB {
			enemy = unit
		}
	}

	if mage == nil || enemy == nil {
		t.Fatal("Could not find mage and enemy")
	}

	// Position them within range
	mage.Position = Position{X: 10, Y: 10}
	enemy.Position = Position{X: 13, Y: 10}

	// Test damage skill
	initialHP = enemy.HP
	affectedUnits, err = state.UseSkill(mage, &enemy.Position)
	assert.NoError(t, err)
	assert.Less(t, enemy.HP, initialHP)
	assert.Equal(t, []int{enemy.ID}, affectedUnits)

	// Test skill out of range
	mage.Position = Position{X: 1, Y: 1}
	enemy.Position = Position{X: 15, Y: 15}
	affectedUnits, err = state.UseSkill(mage, &enemy.Position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
	assert.Nil(t, affectedUnits)
}

func TestUpdateGameState(t *testing.T) {
	state := GetInitialGameState()
	unit := *state.Units[0]

	// Test HOLD action
	affectedUnits, err := state.UpdateGameState(&unit, UnitAction{Action: HOLD, Target: nil}, "")
	assert.NoError(t, err)
	assert.Nil(t, affectedUnits)

	// Test MOVE action
	target := &Position{X: unit.Position.X + 1, Y: unit.Position.Y + 1}
	affectedUnits, err = state.UpdateGameState(&unit, UnitAction{Action: MOVE, Target: target}, "")
	assert.NoError(t, err)
	assert.Equal(t, []int{unit.ID}, affectedUnits)

	// Test multiple MOVE actions (should fail)
	affectedUnits, err = state.UpdateGameState(
		&unit, UnitAction{Action: MOVE, Target: target}, MOVE,
	)
	assert.Error(t, err)
	assert.Nil(t, affectedUnits)

	// Test invalid action
	affectedUnits, err = state.UpdateGameState(
		&unit, UnitAction{Action: "INVALID", Target: target}, "",
	)
	assert.Error(t, err)
	assert.Nil(t, affectedUnits)
}

func TestCheckWinningTeam(t *testing.T) {
	state := GetInitialGameState()

	// Test no winner initially
	teamA := []*Unit{}
	teamB := []*Unit{}
	for _, unit := range state.Units {
		if unit.Team == TeamA {
			teamA = append(teamA, unit)
		} else if unit.Team == TeamB {
			teamB = append(teamB, unit)
		}
	}

	winner, gameOver := checkWinningTeam(teamA, state, teamB)
	assert.False(t, gameOver)
	assert.Equal(t, -1, winner)

	// Test TeamA wins
	for _, unit := range state.Units {
		if unit.Team == TeamB {
			unit.HP = 0
		}
	}

	winner, gameOver = checkWinningTeam(teamA, state, teamB)
	assert.True(t, gameOver)
	assert.Equal(t, TeamA, winner)

	// Test TeamB wins
	for _, unit := range state.Units {
		if unit.Team == TeamA {
			unit.HP = 0
		}
		if unit.Team == TeamB {
			unit.HP = 100
		}
	}

	winner, gameOver = checkWinningTeam(teamA, state, teamB)
	assert.True(t, gameOver)
	assert.Equal(t, TeamB, winner)
}
