package battle

import (
	"aibattle/battler"
	"aibattle/pages"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/samber/lo"

	"github.com/pocketbase/dbx"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type DetailView struct {
	User     *core.Record
	Battle   *core.Record
	Output   string
	MyTeam   string
	Opponent string
}

func Detailed(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")

		battleResult, err := app.FindRecordById("battle_result", id)
		if err != nil {
			return err
		}

		// Load battle relation
		battleErr := app.ExpandRecord(battleResult, []string{"battle", "opponent"}, nil)
		if len(battleErr) > 0 {
			return errors.New("could not load battle data")
		}

		// Get the associated battle record
		battle := battleResult.ExpandedOne("battle")
		if battle == nil {
			return errors.New("battle not found")
		}
		opponent := battleResult.ExpandedOne("opponent")
		if opponent == nil {
			return errors.New("opponent not found")
		}

		outputString := battle.GetString("output")

		reader := bytes.NewReader([]byte(outputString))
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer func(gzReader *gzip.Reader) {
			err := gzReader.Close()
			if err != nil {
				log.Println(err)
			}
		}(gzReader)

		decompressed, err := io.ReadAll(gzReader)
		if err != nil {
			return err
		}
		data := &DetailView{
			User:     e.Auth,
			Battle:   battle,
			Output:   string(decompressed),
			MyTeam:   battleResult.GetString("team"),
			Opponent: opponent.GetString("name"),
		}
		return pages.Render(e, templ, "battle.gohtml", data)
	}
}

type ListView struct {
	ID          string
	ScoreChange string
	Result      string
	Opponent    string
	Date        time.Time
	PromptID    string
}

type ListData struct {
	User    *core.Record
	Battles []ListView
	Error   string
}

func List(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		return defaultList(e, app, templ, "")
	}
}

func defaultList(
	e *core.RequestEvent, app *pocketbase.PocketBase, templ *template.Template, error string,
) error {
	battleViews, battleErr := getUserBattles(e.Auth.Id, app)
	if battleErr != nil {
		return battleErr
	}

	data := &ListData{
		User:    e.Auth,
		Battles: battleViews,
		Error:   error,
	}
	return pages.Render(e, templ, "battle_list.gohtml", data)
}

func getUserBattles(userID string, app *pocketbase.PocketBase) ([]ListView, error) {
	battles := []*core.Record{}
	err := app.RecordQuery("battle_result").
		Join("LEFT JOIN", "users", dbx.NewExp("opponent = users.id")).
		AndWhere(dbx.HashExp{"user": userID}).
		OrderBy("created DESC").
		Limit(100).
		All(&battles)
	if err != nil {
		return nil, err
	}

	// Expand opponent relations to get names
	expErr := app.ExpandRecords(battles, []string{"opponent"}, nil)
	if len(expErr) > 0 {
		return nil, lo.Values(expErr)[0]
	}

	battleViews := make([]ListView, len(battles))
	for i, battle := range battles {
		view := ListView{
			ID:          battle.Id,
			Date:        battle.GetDateTime("created").Time(),
			ScoreChange: fmt.Sprintf("%+.f", battle.GetFloat("score_change")),
			PromptID:    battle.GetString("prompt"),
			Result:      battle.GetString("result"),
		}

		// Get opponent name from expanded relation
		if opponent := battle.ExpandedOne("opponent"); opponent != nil {
			view.Opponent = opponent.GetString("name")
		}

		battleViews[i] = view
	}
	return battleViews, nil
}

// Store last battle time for each user
var lastBattleTime = make(map[string]time.Time)

func RunBattle(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Check rate limit - allow only 1 battle per minute
		userId := e.Auth.Id
		if lastTime, exists := lastBattleTime[userId]; exists {
			elapsed := time.Since(lastTime)
			if elapsed < time.Minute {
				remaining := time.Minute - elapsed
				message := fmt.Sprintf("Please wait %d seconds before starting another battle.", int(remaining.Seconds()))
				return defaultList(e, app, templ, message)
			}
		}

		// Find the active prompt for the current user
		activePrompt, err := app.FindFirstRecordByFilter(
			"prompt",
			"user = {:user} && active = true",
			dbx.Params{
				"user": userId,
			},
		)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				return defaultList(e, app, templ, "You don't have an active prompt. Please activate a prompt before starting a battle.")
			}
			return defaultList(e, app, templ, "Error finding active prompt: "+err.Error())
		}

		// Update the last battle time before running the battle
		lastBattleTime[userId] = time.Now()

		err = battler.RunBattle(app, activePrompt.Id)
		if err != nil {
			return defaultList(e, app, templ, err.Error())
		}

		return e.Redirect(http.StatusFound, "/battle")
	}
}
