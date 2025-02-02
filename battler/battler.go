package battler

import (
	"aibattle/game"
	"aibattle/game/world"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"time"
)

func RunBattleTask(app *pocketbase.PocketBase) {
	for {
		err := RunBattle(app)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(10 * time.Second)
	}
}

func RunBattle(app *pocketbase.PocketBase) error {
	// Find active prompts
	records, err := app.FindRecordsByFilter(
		"prompt",
		"active=TRUE",
		"-created",
		2,
		0,
	)

	if err != nil {
		return fmt.Errorf("error fetching active prompts: %w", err)
	}

	// Need at least 2 prompts for battle
	if len(records) < 2 {
		return errors.New("not enough records")
	}

	prompt1 := records[0]
	prompt2 := records[1]

	// Create battle record
	collection, err := app.FindCollectionByNameOrId("battle")
	if err != nil {
		return fmt.Errorf("error finding battle collection: %w", err)
	}

	battle := core.NewRecord(collection)
	battle.Set("prompt1", prompt1.Id)
	battle.Set("prompt2", prompt2.Id)

	// Run battle
	result, err := GetBattleResult(
		context.Background(),
		prompt1,
		prompt2,
	)

	if err != nil {
		return fmt.Errorf("error running battle: %w", err)
	}

	compressedRes, err := MarshalGzip(result.Turns)
	if err != nil {
		return fmt.Errorf("error comporessing result: %w", err)
	}

	// Save results
	battle.Set("output", compressedRes)
	if result.Winner == world.TeamA {
		battle.Set("winner", "prompt1")
	} else {
		battle.Set("winner", "prompt2")
	}

	if err := app.Save(battle); err != nil {
		return fmt.Errorf("error saving battle: %w", err)
	}
	return nil

}

func MarshalGzip(result []game.TurnLog) (string, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	// Compress output
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return "", fmt.Errorf("error compressing output: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("error closing gzip writer: %w", err)
	}
	return b.String(), nil
}
