package battler

import (
	"aibattle/game"
	"context"
	"fmt"
	"github.com/samber/lo"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func GetBattleResult(
	ctx context.Context, prompt1 *core.Record, prompt2 *core.Record,
) (game.Result, error) {
	fmt.Printf(
		"Run battle team a user: %s prompt %s language %s\n",
		prompt1.GetString("user"), prompt1.Id, prompt1.GetString("language"),
	)
	fmt.Printf(
		"Run battle team b user: %s prompt %s language %s\n",
		prompt2.GetString("user"), prompt2.Id, prompt2.GetString("language"),
	)

	// Get all container IDs
	//killContainers(ctx)

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

	result, err := game.RunGame()
	result = setLogs(ctx, result, env)

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

func killContainers(ctx context.Context) {
	listCmd := exec.CommandContext(ctx, "docker", "ps", "-q")
	containerOutput, listErr := listCmd.Output()
	if listErr != nil {
		log.Printf("Warning: listing containers failed: %v", listErr)
	}
	fmt.Println(string(containerOutput))
	containerIDs := strings.Split(string(containerOutput), "\n")

	for _, containerID := range containerIDs {
		// Kill all running containers
		killCmd := exec.CommandContext(
			ctx, "docker", "kill", containerID,
		)
		killCmd.Stderr = os.Stderr
		killCmd.Stdout = os.Stdout
		if err := killCmd.Run(); err != nil {
			log.Printf("Warning: kill containers failed: %v", err)
		}
	}
}

func setLogs(ctx context.Context, result game.Result, env []string) game.Result {
	teamOneLog, teamOneErr := GetServiceLogs(ctx, "team_one", env)
	if teamOneErr != nil {
		log.Println(teamOneErr)
	}
	teamTwoLog, teamTwoErr := GetServiceLogs(ctx, "team_two", env)
	if teamTwoErr != nil {
		log.Println(teamTwoErr)
	}

	logSize := 1000
	result.TeamOneLogs = lo.Substring(teamOneLog, -logSize, uint(logSize))
	result.TeamTwoLogs = lo.Substring(teamTwoLog, -logSize, uint(logSize))
	return result
}

func GetServiceLogs(ctx context.Context, serviceName string, env []string) (string, error) {
	logs := exec.CommandContext(
		ctx, "docker", "compose",
		"-f", "docker_test/compose.yaml",
		"logs",
		serviceName,
		"--no-color",
	)
	logs.Env = env

	output, err := logs.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for service %s: %v", serviceName, err)
	}

	return string(output), nil
}
