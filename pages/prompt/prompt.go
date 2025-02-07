package prompt

import (
	"aibattle/game/rules"
	"aibattle/pages"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Data struct {
	User *core.Record

	Text   string
	Errors []string

	ID            string
	Output        string
	Status        string
	DefaultPrompt string
	Prompts       []*core.Record
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
		newPrompt, validationErr, promptErr := CreateUpdatePrompt(data.Text, e.Auth.Id, app, nil)
		if promptErr != nil {
			return promptErr
		}
		if validationErr != nil {
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

		updatedPrompt, validationErr, promptErr := CreateUpdatePrompt(
			data.Text, e.Auth.Id, app, prompt,
		)
		if promptErr != nil {
			return promptErr
		}
		// resetting Output field after update
		if validationErr == nil {
			data.Output = updatedPrompt.GetString("output")
			data.Status = updatedPrompt.GetString("status")
		}
		data.Errors = validationErr
		return pages.Render(e, templ, "prompt.gohtml", data)
	}
}

func ActivatePrompt(
	app *pocketbase.PocketBase, templ *template.Template,
) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		prompt, dataErr := app.FindFirstRecordByFilter(
			"prompt", "id={:id} && user={:user} && status='done'",
			dbx.Params{"id": id, "user": e.Auth.Id},
		)
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
	prompts, err := GetPrompts(app, user.Id)
	if err != nil {
		return nil, Data{}, err
	}

	gameRules, err := rules.GetGameDescription()
	if err != nil {
		return nil, Data{}, err
	}
	data := Data{
		User:          user,
		Prompts:       prompts,
		DefaultPrompt: gameRules,
	}

	if id != "" {
		prompt, dataErr := app.FindFirstRecordByFilter(
			"prompt", "id={:id} && user={:user}",
			dbx.Params{"id": id, "user": user.Id},
		)
		if dataErr != nil {
			return nil, data, err
		}
		data.ID = prompt.Id
		data.Text = prompt.GetString("text")
		data.Output = prompt.GetString("output")
		data.Status = prompt.GetString("status")
		promptError := prompt.GetString("error")
		if promptError != "" {
			data.Errors = []string{promptError}
		}
		return prompt, data, nil
	}

	return nil, data, nil
}

var PromptsToProcess = make(chan *core.Record, 20)
var UserRateLimiter = make(map[string]time.Time)

func CreateUpdatePrompt(
	text string, userID string, app *pocketbase.PocketBase, prompt *core.Record,
) (*core.Record, []string, error) {
	// TODO: check prompt limits
	var errors []string
	if len(text) > 300 {
		errors = append(errors, "Text too long")
		return nil, errors, nil
	}
	if time.Now().Sub(UserRateLimiter[userID]).Seconds() < 60 {
		errors = append(errors, "We allow only one update per minute per user, please try later.")
		return nil, errors, nil
	}
	var newPrompt *core.Record
	if prompt == nil {
		collection, err := app.FindCollectionByNameOrId("prompt")
		if err != nil {
			return nil, nil, err
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
		return newPrompt, nil, saveErr
	}

	select {
	case PromptsToProcess <- newPrompt:
		fmt.Println("prompt scheduled", newPrompt.Id)
		UserRateLimiter[userID] = time.Now()
	default:
		return newPrompt, []string{"Too many request to create prompt. Try later."}, nil
	}
	return newPrompt, nil, nil
}
