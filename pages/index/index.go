package index

import (
	"aibattle/pages"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"html/template"
)

type Data struct {
	User *core.Record
}

func Index(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &Data{
			User: e.Auth,
		}
		return pages.Render(e, templ, "index.gohtml", data)
	}
}
