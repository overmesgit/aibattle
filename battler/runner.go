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

	nextTurnFunc, err := PrepareTeams(prompt1.GetString("output"), prompt2.GetString("output"))
	if err != nil {
		return world.Result{}, err
	}
	result, err := world.RunGame(nextTurnFunc)

	if err != nil {
		return world.Result{}, err
	}
	// For now just return placeholder result
	return result, nil
}

func PrepareTeams(team1Text, team2Text string) (func(
	int, world.GameState, int, string,
) (world.UnitAction, error), error) {
	team1FullProg, err := builder.AddGeneratedCodeToTheGameTemplate(team1Text, rules.LangJS)
	if err != nil {
		return nil, err
	}
	team1Action, err := builder.GetGOJAFunction(team1FullProg)
	if err != nil {
		return nil, fmt.Errorf("error preparing js function: %w", err)
	}

	team2FullProg, err := builder.AddGeneratedCodeToTheGameTemplate(team2Text, rules.LangJS)
	if err != nil {
		return nil, err
	}
	team2Action, err := builder.GetGOJAFunction(team2FullProg)
	if err != nil {
		return nil, fmt.Errorf("error preparing js function: %w", err)
	}

	return func(
		team int, state world.GameState, unitID int, actionIndex string,
	) (world.UnitAction, error) {
		var f func(world.GameState, int, string) (world.UnitAction, error)
		switch team {
		case world.TeamA:
			f = team1Action
		case world.TeamB:
			f = team2Action
		default:
			return world.UnitAction{}, fmt.Errorf("wrong team %s", team)
		}
		action, err := f(state, unitID, actionIndex)
		if err != nil {
			return action, fmt.Errorf("error calling GetTurnActions: %w", err)
		}
		return action, nil
	}, nil
}
