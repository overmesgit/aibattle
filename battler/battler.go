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
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RunBattleTask(app *pocketbase.PocketBase) {
	battleTimeoutStr := os.Getenv("BATTLE_TIMEOUT")
	battleTimeout, err := strconv.ParseInt(battleTimeoutStr, 10, 64)
	if err != nil {
		log.Println("BATTLE_TIMEOUT env variable is not a number")
		battleTimeout = 10
	}
	for {
		err := RunBattle(app)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Duration(battleTimeout) * time.Second)
	}
}

func RunBattle(app *pocketbase.PocketBase) error {
	prompt1, prompt2, promptErr := getNextPrompt(app)
	if promptErr != nil {
		return promptErr
	}
	user1Score, user2Score, scoreErr := getScores(
		app, prompt1.GetString("user"), prompt2.GetString("user"),
	)
	if scoreErr != nil {
		return scoreErr
	}
	oldScore1 := user1Score.GetFloat("score")
	oldScore2 := user2Score.GetFloat("score")

	// Run battle
	result, err := GetBattleResult(
		context.Background(),
		prompt1,
		prompt2,
	)
	if err != nil {
		return fmt.Errorf("error running battle: %w", err)
	}

	err = updateUserScores(app, user1Score, user2Score, result)
	if err != nil {
		return fmt.Errorf("error updating scores: %w", err)
	}

	battle, batErr := saveBattle(app, result)
	if batErr != nil {
		return batErr
	}

	resErr := saveBattleResults(
		app, result, user1Score, oldScore1, prompt1, prompt2, battle, user2Score, oldScore2,
	)
	if resErr != nil {
		return resErr
	}

	return nil

}

func saveBattleResults(
	app *pocketbase.PocketBase, result game.Result, user1Score *core.Record, oldScore1 float64,
	prompt1 *core.Record, prompt2 *core.Record, battle *core.Record, user2Score *core.Record,
	oldScore2 float64,
) error {
	user1Res := ""
	user2Res := ""
	switch result.Winner {
	case world.TeamA:
		user1Res = "won"
		user2Res = "lost"
	case world.TeamB:
		user1Res = "lost"
		user2Res = "won"
	case world.Draw:
		user1Res = "draw"
		user2Res = "draw"
	}
	// Create battle result records for both players
	battleResultColl, findErr := app.FindCollectionByNameOrId("battle_result")
	if findErr != nil {
		return fmt.Errorf("error finding battle collection: %w", findErr)
	}
	scoreChange := user1Score.GetFloat("score") - oldScore1
	result1 := createBattleResult(
		prompt1, prompt2.GetString("user"), battle.Id, scoreChange,
		battleResultColl, "teamA", user1Res,
	)
	if res1Err := app.Save(result1); res1Err != nil {
		return fmt.Errorf("error saving battle result 1: %w", res1Err)
	}

	scoreChange2 := user2Score.GetFloat("score") - oldScore2
	result2 := createBattleResult(
		prompt2, prompt1.GetString("user"), battle.Id, scoreChange2,
		battleResultColl, "teamB", user2Res,
	)
	if res2Err := app.Save(result2); res2Err != nil {
		return fmt.Errorf("error saving battle result 2: %w", res2Err)
	}
	return nil
}

func saveBattle(app *pocketbase.PocketBase, result game.Result) (*core.Record, error) {
	compressedRes, zipErr := MarshalGzip(result)
	if zipErr != nil {
		return nil, fmt.Errorf("error comporessing result: %w", zipErr)
	}

	// Create battle record
	collection, colErr := app.FindCollectionByNameOrId("battle")
	if colErr != nil {
		return nil, fmt.Errorf("error finding battle collection: %w", colErr)
	}
	battle := core.NewRecord(collection)
	battle.Set("output", compressedRes)
	if batErr := app.Save(battle); batErr != nil {
		return nil, fmt.Errorf("error saving battle: %w", batErr)
	}
	return battle, nil
}

func getNextPrompt(app *pocketbase.PocketBase) (*core.Record, *core.Record, error) {
	var records []*core.Record
	err := app.RecordQuery("prompt").
		Join("LEFT JOIN", "score", dbx.NewExp("score.user = prompt.user")).
		AndWhere(dbx.HashExp{"prompt.active": true}).
		OrderBy("score.updated ASC").
		Limit(2).
		All(&records)

	if err != nil {
		return nil, nil, fmt.Errorf("error fetching active prompts: %w", err)
	}

	// Need at least 2 prompts for battle
	if len(records) < 2 {
		return nil, nil, errors.New("not enough records")
	}
	prompt1, prompt2 := records[0], records[1]
	return prompt1, prompt2, nil
}

func updateUserScores(
	app *pocketbase.PocketBase, user1Score *core.Record, user2Score *core.Record,
	result game.Result,
) error {
	return app.RunInTransaction(
		func(txApp core.App) error {
			fmt.Printf(
				"user 1 user id: %s score id: %s start score: %f\n", user1Score.GetString("user"),
				user1Score.Id, user1Score.GetFloat("score"),
			)
			fmt.Printf(
				"user 2 user id: %s score id: %s start score: %f\n", user2Score.GetString("user"),
				user2Score.Id, user2Score.GetFloat("score"),
			)
			winner, looser := user1Score, user2Score
			if result.Winner != world.TeamA {
				winner, looser = looser, winner
			}
			newScore1, newScore2 := getNewScores(
				winner.GetFloat("score"), looser.GetFloat("score"), result.Winner == world.Draw,
			)
			fmt.Printf(
				"team %d won, winner %s score %f, looser %s score %f\n", result.Winner, winner.Id,
				newScore1, looser.Id, newScore2,
			)
			winner.Set("score", newScore1)
			looser.Set("score", newScore2)

			// Save both score updates
			if err := txApp.Save(user1Score); err != nil {
				return fmt.Errorf("error updating user1 score: %w", err)
			}
			if err := txApp.Save(user2Score); err != nil {
				return fmt.Errorf("error updating user2 score: %w", err)
			}
			return nil
		},
	)
}

func createBattleResult(
	prompt *core.Record, opponentID string, battleID string, scoreChange float64,
	collection *core.Collection, team string, res string,
) *core.Record {
	result := core.NewRecord(collection)
	result.Set("user", prompt.GetString("user"))
	result.Set("prompt", prompt.Id)
	result.Set("opponent", opponentID)
	result.Set("battle", battleID)
	result.Set("score_change", scoreChange)
	result.Set("team", team)
	result.Set("result", res)
	return result
}

func getScores(
	app *pocketbase.PocketBase, user1 string, user2 string,
) (*core.Record, *core.Record, error) {
	scores := []*core.Record{}
	err := app.RecordQuery("score").
		AndWhere(
			dbx.HashExp{
				"user": dbx.Or(
					dbx.HashExp{"user": user1},
					dbx.HashExp{"user": user2},
				),
			},
		).All(&scores)
	if err != nil {
		return nil, nil, err
	}

	user1Score, user2Score := scores[0], scores[1]
	if user1Score.GetString("user") != user1 {
		user1Score, user2Score = user2Score, user1Score
	}
	return user1Score, user2Score, nil
}

func getNewScores(winner float64, looser float64, draw bool) (float64, float64) {
	// Calculate ELO rating changes
	k := 32.0 // K-factor determines how much ratings can change

	// Calculate expected scores
	e1 := 1.0 / (1.0 + math.Pow(10, (looser-winner)/400.0))
	e2 := 1.0 / (1.0 + math.Pow(10, (winner-looser)/400.0))

	win := float64(1)
	loos := float64(0)
	if draw {
		win = 0.5
		loos = 0.5
	}
	// Update ratings
	newScore1 := math.Round(winner + k*(win-e1))
	newScore2 := math.Round(looser + k*(loos-e2))
	return newScore1, newScore2
}

func MarshalGzip(result game.Result) (string, error) {
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
