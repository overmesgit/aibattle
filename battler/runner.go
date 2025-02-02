package battler

import (
	"aibattle/game"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pocketbase/pocketbase/core"
)

func GetBattleResult(
	ctx context.Context, prompt1 *core.Record, prompt2 *core.Record,
) (game.Result, error) {
	fmt.Printf(
		"Run battle team a %s, team b %s", prompt1.GetString("user"), prompt2.GetString("user"),
	)
	// Set up environment variables for docker-compose
	env := []string{
		"TEAM_ONE=ai" + prompt1.Id,
		"TEAM_TWO=ai" + prompt2.Id,
	}

	// Start docker compose with the environment variables
	compose := exec.CommandContext(
		ctx, "docker", "compose",
		"-f", "docker_test/compose.yaml",
		"up", "--wait",
	)
	compose.Env = env
	compose.Stderr = os.Stderr
	compose.Stdout = os.Stdout

	err := compose.Run()
	if err != nil {
		return game.Result{}, fmt.Errorf("failed to start containers: %v", err)
	}
	// Let containers run for a while
	result, err := game.RunGame()
	if err != nil {
		return game.Result{}, err
	}

	cleanup := exec.CommandContext(
		ctx, "docker", "compose",
		"-f", "docker_test/compose.yaml",
		"down",
	)
	cleanup.Env = env
	cleanup.Stderr = os.Stderr
	cleanup.Stdout = os.Stdout
	err = cleanup.Run()
	if err != nil {
		return game.Result{}, fmt.Errorf("failed to start containers: %v", err)
	}
	// For now just return placeholder result
	return result, nil
}
