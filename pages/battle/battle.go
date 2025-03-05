package battle

import (
	"aibattle/battler"
	"aibattle/pages"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

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
	var battles []*core.Record
	err := app.RecordQuery("battle_result").
		AndWhere(dbx.HashExp{"user": userID}).
		OrderBy("created DESC").
		Limit(100).
		All(&battles)
	if err != nil {
		return nil, err
	}

	userNames, nameErr := getUserNames(battles, app)
	if nameErr != nil {
		return nil, nameErr
	}

	return lo.Map(
		battles, func(b *core.Record, index int) ListView {
			return ListView{
				ID:          b.Id,
				Date:        b.GetDateTime("created").Time(),
				ScoreChange: fmt.Sprintf("%+.f", b.GetFloat("score_change")),
				PromptID:    b.GetString("prompt"),
				Result:      b.GetString("result"),
				Opponent:    userNames[b.GetString("opponent")],
			}
		},
	), nil
}

func getUserNames(
	battles []*core.Record, app *pocketbase.PocketBase,
) (map[string]string, error) {
	opponentIDs := lo.Uniq(
		lo.Map(
			battles, func(b *core.Record, index int) string {
				return b.GetString("opponent")
			},
		),
	)
	opponents, err := app.FindRecordsByIds("users", opponentIDs)
	if err != nil {
		return nil, err
	}
	userNames := lo.SliceToMap(
		opponents, func(o *core.Record) (string, string) {
			return o.Id, o.GetString("name")
		},
	)
	return userNames, nil
}

// Store last battle time for each user
var lastBattleTime = make(map[string]time.Time)

func RunBattle(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Check rate limit - allow only 1 battle per minute
		userId := e.Auth.Id
		limitError := checkBattleLimit(userId)
		if limitError != nil {
			return defaultList(e, app, templ, limitError.Error())
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
				return defaultList(
					e, app, templ,
					"You don't have an active prompt. Please activate a prompt before starting a battle.",
				)
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

func checkBattleLimit(userId string) error {
	if lastTime, exists := lastBattleTime[userId]; exists {
		elapsed := time.Since(lastTime)
		if elapsed < time.Minute {
			remaining := time.Minute - elapsed
			return fmt.Errorf(
				"please wait %d seconds before starting another battle", int(remaining.Seconds()),
			)
		}
	}
	return nil
}
