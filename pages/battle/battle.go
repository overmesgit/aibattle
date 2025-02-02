package battle

import (
	"aibattle/pages"
	"bytes"
	"compress/gzip"
	"html/template"
	"io"
	"log"

	"github.com/pocketbase/pocketbase/tools/types"

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

		battle, err := app.FindRecordById("battle", id)
		if err != nil {
			return err
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
	Date        types.DateTime
}

type IndexData struct {
	User    *core.Record
	Battles []BattleView
}

func Index(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {

		// Get all battles for the current user
		battles, err := app.FindRecordsByFilter(
			"battle_result",
			"user = {:userId}",
			"-created",
			100,
			0,
			dbx.Params{"userId": e.Auth.Id},
		)
		if err != nil {
			return err
		}

		battleViews := make([]BattleView, len(battles))
		for i, battle := range battles {
			view := BattleView{
				ID:   battle.Id,
				Date: battle.GetDateTime("created"),
			}
			view.ScoreChange = battle.GetFloat("score_change")
			if view.ScoreChange > 0 {
				view.Status = "won"
			} else {
				view.Status = "lost"
			}

			battleViews[i] = view
		}

		data := &IndexData{
			User:    e.Auth,
			Battles: battleViews,
		}
		return pages.Render(e, templ, "battle_list.gohtml", data)
	}
}
