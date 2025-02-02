package leader

import (
	"aibattle/pages"
	"github.com/pocketbase/dbx"
	"github.com/samber/lo"
	"html/template"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type LeaderData struct {
	Scores []ScoreEntry
	User   *core.Record
}

type ScoreEntry struct {
	Username string
	UserID   string
	Score    float64
}

func List(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get scores ordered by score descending
		var records []*core.Record
		err := app.RecordQuery("score").
			Join("LEFT JOIN", "users", dbx.NewExp("users.id = score.user")).
			OrderBy("score.score DESC").
			All(&records)

		if err != nil {
			return err
		}

		// Expand opponent relations to get names
		expErr := app.ExpandRecords(records, []string{"user"}, nil)
		if len(expErr) > 0 {
			return lo.Values(expErr)[0]
		}

		// Build score entries
		scores := make([]ScoreEntry, 0)
		for _, record := range records {
			user := record.ExpandedOne("user")
			if user != nil {
				scores = append(
					scores, ScoreEntry{
						Username: user.GetString("name"),
						UserID:   user.Id,
						Score:    record.GetFloat("score"),
					},
				)
			}
		}

		data := &LeaderData{
			Scores: scores,
			User:   e.Auth,
		}

		return pages.Render(e, templ, "leader.gohtml", data)
	}
}
