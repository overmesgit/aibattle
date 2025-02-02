package auth

import (
	"aibattle/pages"
	"errors"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"html/template"
	"net/http"
	"time"
)

type LoginData struct {
	Email string
	Error string
	User  *core.Record
}

func Login(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		data := &SignUpData{
			User: e.Auth,
		}

		if e.Request.Method == "POST" {
			email := e.Request.FormValue("email")
			password := e.Request.FormValue("password")
			user, valErr := validateUser(app, email, data, password)
			if valErr != nil {
				data.Error = valErr.Error()
			} else {
				expiration := time.Now().Add(3 * 24 * time.Hour)
				token, err := user.NewAuthToken()
				if err != nil {
					return err
				}
				cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
				http.SetCookie(e.Response, &cookie)
				return e.Redirect(http.StatusFound, "/")
			}
		}
		return pages.Render(e, templ, "login.gohtml", data)
	}
}

func validateUser(
	app *pocketbase.PocketBase, email string, data *SignUpData, password string,
) (*core.Record, error) {
	user, err := app.FindAuthRecordByEmail("users", email)
	if err != nil {
		return user, errors.New("wrong user name or password")
	}

	if !user.ValidatePassword(password) {
		return user, errors.New("wrong user name or password")
	}
	return user, nil
}
