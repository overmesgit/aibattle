package rules

import (
	"aibattle/game/world"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/samber/lo"
)

const (
	LangPy = "py"
	LangGo = "go"
	LangJS = "js"
)

var AvailableLanguages = []string{LangJS}

func GetGameDescription(language string) (string, error) {
	state := world.GetInitialGameState()
	var unitsDescription strings.Builder
	uniqueUnits := lo.Filter(
		state.Units, func(unit *world.Unit, index int) bool {
			return index%2 == 0
		},
	)
	for _, unit := range uniqueUnits {
		unitActions := state.UnitActionMap[unit.Type]
		actions := printActions(unitActions)
		unitsDescription.WriteString(
			fmt.Sprintf(
				"Unit: type %s, initiative %d, hp %d, actions %s\n", unit.Type, unit.Initiative,
				unit.MaxHP,
				actions,
			),
		)
	}

	gameStateJson, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", err
	}

	nextActionExample, err := json.Marshal(
		world.UnitAction{Action: world.MOVE, Target: &world.Position{X: 20, Y: 20}},
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
		GameState         string
		NextActionExample string
		LanguageTemplate  string
	}{
		NumUnitsPerTeam:   len(state.Units) / 2,
		GridSize:          fmt.Sprintf("%dx%d", state.Height, state.Width),
		UnitsDescription:  unitsDescription.String(),
		GameState:         string(gameStateJson),
		NextActionExample: string(nextActionExample),
		LanguageTemplate:  languageTemplate,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

//go:embed templates/py.py templates/go_test.go templates/js.js
var templateFS embed.FS

func getLanguageTemplate(language string) (string, error) {
	var templatePath string
	switch language {
	case LangPy:
		templatePath = "templates/py.py"
	case LangGo:
		templatePath = "templates/go_test.go"
	case LangJS:
		templatePath = "templates/js.js"
	default:
		return "", errors.New("unknown language")
	}

	// Read the template content from the embedded file system
	templateContent, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse and execute the template
	templateString := string(templateContent)
	log.Printf("Language template length: %d", len(templateString))
	return templateString, nil
}

func AddGeneratedCodeToTheGameTemplate(generatedProg string, language string) (string, error) {
	languageTemplate, err := getLanguageTemplate(language)
	if err != nil {
		return "", err
	}
	return languageTemplate + generatedProg, nil
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
