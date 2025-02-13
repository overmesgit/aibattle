package battler

import (
	"errors"
	"fmt"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"os"
	"strconv"
	"time"
)

var BattleChannel = make(chan string, 100)

func RunBattleTask(app *pocketbase.PocketBase) {
	go func() {
		for {
			nextPromptID := <-BattleChannel
			err := RunBattle(app, nextPromptID)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	minSleepDuration := 10 * time.Second
	targetBattlesPerUser := GetTimeBetweenBattles()
	for {
		activePrompts, err := GetNumberOfActivePrompts(app, minSleepDuration)
		if err != nil {
			log.Println(err)
			time.Sleep(minSleepDuration)
			continue
		}

		// Calculate sleep duration to achieve ~1 battle per user per 5 minutes
		sleepDuration := targetBattlesPerUser / time.Duration(activePrompts)
		BattleChannel <- ""
		time.Sleep(sleepDuration)
	}
}

func GetNumberOfActivePrompts(
	app *pocketbase.PocketBase, minSleepDuration time.Duration,
) (int, error) {
	// Count active prompts/users
	var records []*core.Record
	err := app.RecordQuery("prompt").
		AndWhere(dbx.HashExp{"active": true}).
		All(&records)

	if err != nil {
		return 0, fmt.Errorf("error getting active prompt count: %w", err)
	}

	activePrompts := len(records)
	if activePrompts < 2 {
		return 0, errors.New("not enough active prompts")
	}
	return activePrompts, nil
}

func GetTimeBetweenBattles() time.Duration {
	timeBetween := os.Getenv("TIME_BETWEEN_BATTLES_MIN")
	parseInt, err := strconv.ParseInt(timeBetween, 10, 64)
	// default is 3 minutes if env isn't set
	targetBattlesPerUser := 3 * time.Minute
	if err == nil {
		targetBattlesPerUser = time.Duration(parseInt) * time.Minute
	} else {
		fmt.Println("TIME_BETWEEN_BATTLES_MIN", err)
	}
	return targetBattlesPerUser
}
