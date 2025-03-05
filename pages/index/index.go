package index

import (
	"aibattle/pages"
	"html/template"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Data struct {
	User *core.Record
}

func Index(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &Data{
			User: e.Auth,
		}
		return pages.Render(e, templ, "index/index.gohtml", data)
	}
}

func Landing(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &Data{
			User: e.Auth,
		}
		return pages.Render(e, templ, "index/landing.gohtml", data)
	}
}
