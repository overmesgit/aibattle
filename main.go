package main

import (
	_ "aibattle/migrations"
	"aibattle/pages"
	"aibattle/pages/auth"
	"aibattle/pages/battle"
	"aibattle/pages/index"
	"aibattle/pages/leader"
	"aibattle/pages/middleware"
	"aibattle/pages/prompt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/router"
)

func main() {
	app := pocketbase.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(
		app, app.RootCmd, migratecmd.Config{
			Automigrate: isGoRun,
		},
	)
	templ, err := pages.ParseTemplates()
	if err != nil {
		log.Fatal(err)
	}

	app.OnServe().BindFunc(
		func(se *core.ServeEvent) error {
			se.Router.Bind(apis.Gzip())
			se.Router.Bind(middleware.LoadAuthToken())
			se.Router.GET(
				"/dist/{path...}", apis.Static(os.DirFS("./dist"), false),
			).Bind(apis.Gzip())
			se.Router.GET("/signup", auth.SignUp(app, templ))
			se.Router.POST("/signup", auth.SignUp(app, templ))
			se.Router.GET("/login", auth.Login(app, templ))
			se.Router.POST("/login", auth.Login(app, templ))
			se.Router.GET("/leader", leader.List(app, templ))

			se.Router.GET("/{$}", index.Landing(app, templ))

			withAuth(
				se.Router.GET("/logout", auth.Logout(app, templ)),
				se.Router.GET("/prompt", prompt.NewPromptForm(app, templ)),
				se.Router.POST("/prompt", prompt.CreatePrompt(app, templ)),
				se.Router.GET("/prompt/{id}", prompt.DetailedPrompt(app, templ)),
				se.Router.POST("/prompt/{id}", prompt.UpdatePrompt(app, templ)),
				se.Router.POST("/prompt/{id}/activate", prompt.ActivatePrompt(app)),
				se.Router.GET("/battle", battle.List(app, templ)),
				se.Router.GET("/battle/{id}", battle.Detailed(app, templ)),
				se.Router.POST("/battle/run", battle.RunBattle(app, templ)),
			)

			go func() {
				prompt.ProcessPrompts(app)
			}()
			return se.Next()
		},
	)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func withAuth(routes ...*router.Route[*core.RequestEvent]) {
	for _, r := range routes {
		r.
			BindFunc(
				func(e *core.RequestEvent) error {
					if e.Auth == nil {
						return e.Redirect(http.StatusFound, "/login")
					}
					return e.Next()
				},
			).
			Bind(apis.RequireAuth())

	}
}
