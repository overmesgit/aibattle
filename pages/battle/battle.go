package battle

import (
	"aibattle/pages"
	"bytes"
	"compress/gzip"
	"errors"
	"html/template"
	"io"
	"log"
	"time"

	"github.com/samber/lo"

	"github.com/pocketbase/dbx"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Data struct {
	User   *core.Record
	Battle *core.Record
	Output string
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
		battleErr := app.ExpandRecord(battleResult, []string{"battle"}, nil)
		if len(battleErr) > 0 {
			return errors.New("could not load battle data")
		}

		// Get the associated battle record
		battle := battleResult.ExpandedOne("battle")
		if battle == nil {
			return errors.New("battle not found")
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
		output := string(decompressed)
		data := &Data{
			User:   e.Auth,
			Battle: battle,
			Output: output,
		}
		return pages.Render(e, templ, "battle.gohtml", data)
	}
}

type BattleView struct {
	ID          string
	Status      string
	ScoreChange float64
	Opponent    string
	Date        time.Time
}

type ListData struct {
	User    *core.Record
	Battles []BattleView
}

func List(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {

		battles := []*core.Record{}
		err := app.RecordQuery("battle_result").
			Join("LEFT JOIN", "users", dbx.NewExp("opponent = users.id")).
			AndWhere(dbx.HashExp{"user": e.Auth.Id}).
			OrderBy("created DESC").
			Limit(100).
			All(&battles)
		if err != nil {
			return err
		}

		// Expand opponent relations to get names
		expErr := app.ExpandRecords(battles, []string{"opponent"}, nil)
		if len(expErr) > 0 {
			return lo.Values(expErr)[0]
		}

		battleViews := make([]BattleView, len(battles))
		for i, battle := range battles {
			view := BattleView{
				ID:          battle.Id,
				Date:        battle.GetDateTime("created").Time(),
				ScoreChange: battle.GetFloat("score_change"),
			}
			if view.ScoreChange > 0 {
				view.Status = "won"
			} else {
				view.Status = "lost"
			}

			// Get opponent name from expanded relation
			if opponent := battle.ExpandedOne("opponent"); opponent != nil {
				view.Opponent = opponent.GetString("name")
			}

			battleViews[i] = view
		}

		data := &ListData{
			User:    e.Auth,
			Battles: battleViews,
		}
		return pages.Render(e, templ, "battle_list.gohtml", data)
	}
}
