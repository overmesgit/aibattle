package auth

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"html/template"
	"net/http"
	"time"
)

func Logout(app *pocketbase.PocketBase, templ *template.Template) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		expiration := time.Now()
		cookie := http.Cookie{Name: "token", Value: "", Expires: expiration}
		http.SetCookie(e.Response, &cookie)
		return e.Redirect(http.StatusFound, "/")
	}
}
