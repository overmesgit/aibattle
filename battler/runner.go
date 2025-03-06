package battler

import (
	"aibattle/game/rules"
	"aibattle/game/world"
	"aibattle/pages/builder"
	"context"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func GetBattleResult(
	ctx context.Context, prompt1 *core.Record, prompt2 *core.Record,
) (world.Result, error) {
	fmt.Printf(
		"Run battle team a user: %s prompt %s language %s\n",
		prompt1.GetString("user"), prompt1.Id, prompt1.GetString("language"),
	)
	fmt.Printf(
		"Run battle team b user: %s prompt %s language %s\n",
		prompt2.GetString("user"), prompt2.Id, prompt2.GetString("language"),
	)

	match, err := NewMatch(prompt1.GetString("output"), prompt2.GetString("output"))
	if err != nil {
		return world.Result{}, err
	}
	result, err := world.RunGame(match.GetTeamNextAction)

	if err != nil {
		return world.Result{}, err
	}
	// For now just return placeholder result
	return result, nil
}

type Match struct {
	teamOne builder.GOJARunner
	teamTwo builder.GOJARunner
}

func NewMatch(team1Text, team2Text string) (Match, error) {
	team1FullProg, err := rules.AddGeneratedCodeToTheGameTemplate(team1Text, rules.LangJS)
	if err != nil {
		return Match{}, err
	}
	team1Action, err := builder.NewGOJARunner(team1FullProg)
	if err != nil {
		return Match{}, fmt.Errorf("error preparing js function: %w", err)
	}

	team2FullProg, err := rules.AddGeneratedCodeToTheGameTemplate(team2Text, rules.LangJS)
	if err != nil {
		return Match{}, err
	}
	team2Action, err := builder.NewGOJARunner(team2FullProg)
	if err != nil {
		return Match{}, fmt.Errorf("error preparing js function: %w", err)
	}

	return Match{
		teamOne: team1Action,
		teamTwo: team2Action,
	}, nil
}

func (m Match) GetTeamNextAction(
	team int, state world.GameState, unitID int, actionIndex string,
) (world.UnitAction, error) {
	var runner builder.GOJARunner
	switch team {
	case world.TeamA:
		runner = m.teamOne
	case world.TeamB:
		runner = m.teamTwo
	default:
		return world.UnitAction{}, fmt.Errorf("wrong team %s", team)
	}
	action, err := runner.GetNextAction(state, unitID, actionIndex)
	if err != nil {
		return action, fmt.Errorf("error calling GetTurnActions: %w", err)
	}
	return action, nil
}
