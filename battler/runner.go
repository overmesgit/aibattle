package battler

import (
	"aibattle/game/world"
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

	result, err := world.RunGame(Prepare(prompt1.GetString("output"), prompt2.GetString("output")))

	if err != nil {
		return world.Result{}, err
	}
	// For now just return placeholder result
	return result, nil
}

func Prepare(team1Text, team2Text string) func(
	world.Team, world.GameState, int, world.ActionIndex,
) (world.UnitAction, error) {
	return func(
		team world.Team, state world.GameState, unitID int, actionIndex world.ActionIndex,
	) (world.UnitAction, error) {

		return world.UnitAction{}, nil
	}
}
