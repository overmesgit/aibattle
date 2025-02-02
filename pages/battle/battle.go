package battle

import (
	"aibattle/pages"
	"github.com/pocketbase/pocketbase/tools/types"
	"html/template"

	"github.com/pocketbase/dbx"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Data struct {
	User *core.Record
}

func Detailed(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &Data{
			User: e.Auth,
		}
		return pages.Render(e, templ, "battle.gohtml", data)
	}
}

type BattleView struct {
	ID     string
	Status string
	Date   types.DateTime
}

type IndexData struct {
	User    *core.Record
	Battles []BattleView
}

func Index(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {

		// Get all prompts for the current user
		prompts, err := app.FindRecordsByFilter(
			"prompt",
			"user = {:userId}",
			"-created",
			1000,
			0,
			dbx.Params{"userId": e.Auth.Id},
		)
		if err != nil {
			return err
		}

		// Create a set of prompt IDs
		promptIds := make(map[string]struct{})
		for _, prompt := range prompts {
			promptIds[prompt.Id] = struct{}{}
		}

		// Get all battles for the current user
		battles, err := app.FindRecordsByFilter(
			"battle",
			"prompt1.user = {:userId} || prompt2.user = {:userId}",
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
			winner := battle.GetString("winner")
			myPrompt1 := promptIds[battle.GetString("prompt1")] == struct{}{}
			myPrompt2 := promptIds[battle.GetString("prompt2")] == struct{}{}
			if (winner == "prompt1" && myPrompt1) ||
				(winner == "prompt2" && myPrompt2) {
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
