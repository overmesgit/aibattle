package auth

import (
	"aibattle/pages"
	"html/template"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type SignUpData struct {
	Email string
	Error string
	User  *core.Record
}

func SignUp(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &SignUpData{
			User: e.Auth,
		}

		if e.Request.Method == "POST" {
			email := e.Request.FormValue("email")
			password := e.Request.FormValue("password")
			username := e.Request.FormValue("name")
			collection, err := app.FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}

			newUser := core.NewRecord(collection)
			newUser.SetEmail(email)
			newUser.SetPassword(password)
			newUser.Set("name", username)

			saveErr := app.Save(newUser)
			if saveErr != nil {
				data.Error = saveErr.Error()
			} else {
				scoreCollection, err := app.FindCollectionByNameOrId("score")
				if err != nil {
					return err
				}

				score := core.NewRecord(scoreCollection)
				score.Set("score", 1000)
				score.Set("user", newUser.Id)

				if saveErr := app.Save(score); saveErr != nil {
					return saveErr
				}
				return redirectWithCookie(e, newUser)
			}
		}
		return pages.Render(e, templ, "signup.gohtml", data)
	}
}

func redirectWithCookie(e *core.RequestEvent, newUser *core.Record) error {
	expiration := time.Now().Add(3 * 24 * time.Hour)
	token, err := newUser.NewAuthToken()
	if err != nil {
		return err
	}
	cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
	http.SetCookie(e.Response, &cookie)
	return e.Redirect(http.StatusFound, "/")
}
