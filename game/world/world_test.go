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
	err := state.MoveUnit(unit, target)
	assert.NoError(t, err)
	assert.Equal(t, *target, unit.Position)

	// Test move out of range
	target = &Position{X: originalPos.X + 10, Y: originalPos.Y + 10}
	err = state.MoveUnit(unit, target)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of moving range")

	// Test move out of map
	target = &Position{X: -1, Y: -1}
	err = state.MoveUnit(unit, target)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of map range")

	// Test move to occupied position
	occupiedPos := &Position{X: state.Units[2].Position.X, Y: state.Units[2].Position.Y}
	err = state.MoveUnit(unit, occupiedPos)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target is occupied")
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
	err := state.AttackUnit(*attacker, &targetUnit.Position)
	assert.NoError(t, err)
	assert.Less(t, targetUnit.HP, initialHP)

	// Test attack out of range
	attacker.Position = Position{X: 1, Y: 1}
	targetUnit.Position = Position{X: 10, Y: 10}
	err = state.AttackUnit(*attacker, &targetUnit.Position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")

	// Test attack non-existent unit
	emptyPos := &Position{X: 15, Y: 15}
	err = state.AttackUnit(*attacker, emptyPos)
	assert.Error(t, err)
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
	err := state.UseSkill(*healer, &teammate.Position)
	assert.NoError(t, err)
	assert.Greater(t, teammate.HP, initialHP)

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
	err = state.UseSkill(*mage, &enemy.Position)
	assert.NoError(t, err)
	assert.Less(t, enemy.HP, initialHP)

	// Test skill out of range
	mage.Position = Position{X: 1, Y: 1}
	enemy.Position = Position{X: 15, Y: 15}
	err = state.UseSkill(*mage, &enemy.Position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestUpdateGameState(t *testing.T) {
	state := GetInitialGameState()
	unit := *state.Units[0]

	// Test HOLD action
	err := state.UpdateGameState(unit, UnitAction{Action: HOLD, Target: nil}, "")
	assert.NoError(t, err)

	// Test MOVE action
	target := &Position{X: unit.Position.X + 1, Y: unit.Position.Y + 1}
	err = state.UpdateGameState(unit, UnitAction{Action: MOVE, Target: target}, "")
	assert.NoError(t, err)

	// Test multiple MOVE actions (should fail)
	err = state.UpdateGameState(unit, UnitAction{Action: MOVE, Target: target}, MOVE)
	assert.Error(t, err)

	// Test invalid action
	err = state.UpdateGameState(unit, UnitAction{Action: "INVALID", Target: target}, "")
	assert.Error(t, err)
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
	assert.Equal(t, Team(-1), winner)

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
