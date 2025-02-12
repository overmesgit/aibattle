package rules

import (
	"aibattle/game"
	"aibattle/game/world"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

const (
	LangPy = "py"
	LangGo = "go"
)

var AvailableLanguages = []string{LangPy, LangGo}

func GetGameDescription(language string) (string, error) {
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

	jsonState, err := json.MarshalIndent(game.NextTurnInput{state, 1}, "", "  ")
	if err != nil {
		return "", err
	}

	nextActionExample, err := json.Marshal(
		game.TurnAction{[2]*game.UnitAction{
			{Action: world.MOVE, Target: &world.Position{X: 20, Y: 20}},
			{Action: world.ATTACK1, Target: &world.Position{X: 21, Y: 20}},
		}, 10},
	)
	if err != nil {
		return "", err
	}

	languageTemplate, langErr := getLanguageTemplate(language)
	if langErr != nil {
		return "", langErr
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
		LanguageTemplate  string
	}{
		NumUnitsPerTeam:   len(state.Units) / 2,
		GridSize:          fmt.Sprintf("%dx%d", state.Height, state.Width),
		UnitsDescription:  unitsDescription.String(),
		JSONState:         string(jsonState),
		NextActionExample: string(nextActionExample),
		LanguageTemplate:  languageTemplate,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getLanguageTemplate(language string) (string, error) {
	languageTemplate := ""
	switch language {
	case LangPy:
		languageTemplate = "game/rules/templates/py.py"
	case LangGo:
		languageTemplate = "game/rules/templates/go_test.go"
	default:
		return "", errors.New("unknown language")
	}

	langTemplate, langErr := template.ParseFiles(languageTemplate)
	if langErr != nil {
		return "", langErr
	}
	var langBuf bytes.Buffer
	if err := langTemplate.Execute(&langBuf, nil); err != nil {
		return "", err
	}
	return langBuf.String(), nil
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
