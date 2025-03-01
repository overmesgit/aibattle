package prompt

import (
	"aibattle/battler"
	"aibattle/pages/builder"
	"context"
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func ProcessPrompts(app *pocketbase.PocketBase) {
	ScheduleRemainingPrompts(app)
	for {
		nextPrompt := <-PromptsToProcess
		newProg, promptErr := builder.GetProgram(
			context.Background(), nextPrompt.GetString("text"),
			nextPrompt.GetString("language"),
		)
		if promptErr != nil {
			log.Printf("Error getting prompt: %v", promptErr)
			nextPrompt.Set("status", "error")
			nextPrompt.Set("error", promptErr.Error())
		} else {
			nextPrompt.Set("status", "done")
			nextPrompt.Set("error", "")
		}
		nextPrompt.Set("output", newProg)
		saveErr := app.Save(nextPrompt)
		if saveErr != nil {
			log.Printf("Error saving prompt: %v", saveErr)
		}
		activateIfFirstPrompt(app, nextPrompt)
	}
}

func activateIfFirstPrompt(app *pocketbase.PocketBase, prompt *core.Record) error {
	if prompt.GetBool("active") {
		return nil
	}
	user := prompt.GetString("user")
	records, err := app.FindRecordsByFilter(
		"prompt",
		"user = {:user} && active = true",
		"-created",
		1,
		0,
		dbx.Params{
			"user": user,
		},
	)
	if err != nil {
		log.Printf("Error checking active prompts: %v", err)
		return err
	}
	if len(records) == 0 {
		prompt.Set("active", true)
		saveError := app.Save(prompt)
		if saveError != nil {
			log.Printf("Error activating prompt: %v", saveError)
			return saveError
		}

		battler.BattleChannel <- prompt.Id
	}
	return nil
}

func ScheduleRemainingPrompts(app *pocketbase.PocketBase) {
	records, err := app.FindRecordsByFilter(
		"prompt",
		"status = ''",
		"-created",
		20,
		0,
	)
	if err != nil {
		log.Fatalf("Error fetching prompt: %v", err)
	}
	for _, record := range records {
		log.Printf("Scheduling record: %v", record.Id)
		PromptsToProcess <- record
	}
}
