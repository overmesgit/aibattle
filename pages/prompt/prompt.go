package prompt

import (
	"aibattle/pages"
	"html/template"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Data struct {
	User *core.Record

	Text   string
	Errors []string

	ID      string
	Output  string
	Prompts []*core.Record
}

func GetPrompts(app *pocketbase.PocketBase, userId string) ([]*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"prompt",
		"user = {:user}",
		"-created",
		0,
		0,
		dbx.Params{
			"user": userId,
		},
	)
	return records, err
}

func NewPromptForm(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, data, dataErr := defaultData(e.Auth, app, "")
		if dataErr != nil {
			return dataErr
		}
		return pages.Render(e, templ, "prompt.gohtml", data)
	}
}

func CreatePrompt(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, data, dataErr := defaultData(e.Auth, app, "")
		if dataErr != nil {
			return dataErr
		}
		data.Text = e.Request.FormValue("text")
		newPrompt, promptErr := CreateUpdatePrompt(data.Text, e.Auth.Id, app, nil)
		if promptErr != nil {
			return pages.Render(e, templ, "prompt.gohtml", data)
		}
		return e.Redirect(http.StatusFound, "/prompt/"+newPrompt.Id)
	}
}

func DetailedPrompt(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		_, data, dataErr := defaultData(e.Auth, app, id)
		if dataErr != nil {
			return dataErr
		}
		return pages.Render(e, templ, "prompt.gohtml", data)
	}
}

func UpdatePrompt(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		prompt, data, dataErr := defaultData(e.Auth, app, id)
		if dataErr != nil {
			return dataErr
		}

		data.Text = e.Request.FormValue("text")

		updatedPrompt, promptErr := CreateUpdatePrompt(data.Text, e.Auth.Id, app, prompt)
		data.Output = updatedPrompt.GetString("output")
		data.Errors = promptErr
		return pages.Render(e, templ, "prompt.gohtml", data)
	}
}

func ActivatePrompt(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		prompt, _, dataErr := defaultData(e.Auth, app, id)
		if dataErr != nil {
			return dataErr
		}
		// set all other prompts to inactive first
		_, err := app.DB().NewQuery("UPDATE prompt SET active = FALSE WHERE user = {:user}").
			Bind(
				dbx.Params{
					"user": e.Auth.Id,
				},
			).
			Execute()
		if err != nil {
			return err
		}

		prompt.Set("active", true)
		saveError := app.Save(prompt)
		if saveError != nil {
			return saveError
		}
		return e.Redirect(http.StatusFound, "/prompt/"+prompt.Id)
	}
}

func defaultData(
	user *core.Record, app *pocketbase.PocketBase, id string,
) (*core.Record, Data, error) {
	// TODO: check user
	prompts, err := GetPrompts(app, user.Id)
	if err != nil {
		return nil, Data{}, err
	}

	data := Data{
		User:    user,
		Prompts: prompts,
	}

	if id != "" {
		prompt, err := app.FindRecordById("prompt", id)
		if err != nil {
			return nil, data, err
		}
		data.ID = prompt.Id
		data.Text = prompt.GetString("text")
		data.Output = prompt.GetString("output")
		promptError := prompt.GetString("error")
		if promptError != "" {
			data.Errors = []string{promptError}
		}
		return prompt, data, nil
	}

	return nil, data, nil
}

func CreateUpdatePrompt(
	text string, userID string, app *pocketbase.PocketBase, prompt *core.Record,
) (*core.Record, []string) {
	// TODO: check prompt limits
	var errors []string
	if len(text) > 300 {
		errors = append(errors, "Text too long")
		return nil, errors
	}
	var newPrompt *core.Record
	if prompt == nil {
		collection, err := app.FindCollectionByNameOrId("prompt")
		if err != nil {
			return nil, []string{err.Error()}
		}
		newPrompt = core.NewRecord(collection)
	} else {
		newPrompt = prompt
	}
	newPrompt.Set("text", text)
	newPrompt.Set("user", userID)
	newPrompt.Set("status", "")
	newPrompt.Set("output", "")
	saveErr := app.Save(newPrompt)
	if saveErr != nil {
		return nil, []string{saveErr.Error()}
	}
	return newPrompt, nil
}
