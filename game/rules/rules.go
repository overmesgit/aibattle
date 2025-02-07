package rules

import (
	"aibattle/game/world"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

type TurnAction struct {
	UnitAction [2]UnitAction `json:"unit_action"`
}

type UnitAction struct {
	Action world.Action    `json:"action"`
	Target *world.Position `json:"target"`
}

type NextTurnInput struct {
	State  world.GameState `json:"state"`
	UnitID int             `json:"unit_id"`
}

func GetGameDescription() (string, error) {
	state := world.GetInitialGameState()
	unitsList := []*world.Unit{
		world.NewWarrior(world.TeamA, world.Position{X: 20, Y: 20}),
		world.NewHealer(world.TeamA, world.Position{X: 20, Y: 20}),
		world.NewMage(world.TeamA, world.Position{X: 20, Y: 20}),
		world.NewRogue(world.TeamA, world.Position{X: 20, Y: 20}),
	}
	var unitsDescription strings.Builder
	for _, unit := range unitsList {
		actions := printActions(unit.Actions)
		unitsDescription.WriteString(
			fmt.Sprintf(
				"Unit: type %s, initiative %d, hp %d, actions %s\n", unit.Type, unit.Initiative,
				unit.MaxHP,
				actions,
			),
		)
	}
	possibleActions := fmt.Sprint(
		world.HOLD, ", ",
		world.MOVE, ", ",
		world.ATTACK1, ", ",
		world.ATTACK2, ", ",
		world.SKILL1, ", ",
		world.SKILL2, ", ",
	)

	jsonState, err := json.MarshalIndent(NextTurnInput{state, 1}, "", "  ")
	if err != nil {
		return "", err
	}

	nextActionExample, err := json.Marshal(
		TurnAction{[2]UnitAction{
			{Action: world.MOVE, Target: &world.Position{X: 20, Y: 20}},
			{Action: world.ATTACK1, Target: &world.Position{X: 21, Y: 20}},
		}},
	)
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles("game/rules/rules.tmpl")
	if err != nil {
		return "", err
	}

	data := struct {
		NumUnitsPerTeam   int
		GridSize          string
		UnitsDescription  string
		JSONState         string
		NextActionExample string
		PossibleActions   string
	}{
		NumUnitsPerTeam:   len(state.Units) / 2,
		GridSize:          fmt.Sprintf("%dx%d", state.Height, state.Width),
		UnitsDescription:  unitsDescription.String(),
		JSONState:         string(jsonState),
		NextActionExample: string(nextActionExample),
		PossibleActions:   possibleActions,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func printActions(actions world.ActionMap) string {
	var sb strings.Builder

	if actions.Hold != nil {
		sb.WriteString(
			fmt.Sprintf("hold"),
		)
	}

	if actions.Move != nil {
		sb.WriteString(fmt.Sprintf(", move range %d", actions.Move.Distance))
	}

	if actions.Attack1 != nil {
		sb.WriteString(
			fmt.Sprintf(
				", attack1 range %d damage %d", actions.Attack1.Range, actions.Attack1.Damage,
			),
		)
	}

	if actions.Skill1 != nil {
		sb.WriteString(
			fmt.Sprintf(
				", skill1 effect %s range %d value %d", actions.Skill1.Effect, actions.Skill1.Range,
				actions.Skill1.Value,
			),
		)
	}

	return sb.String()
}
