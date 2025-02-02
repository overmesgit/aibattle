package main

import (
	"aibattle/battler"
	_ "aibattle/migrations"
	"aibattle/pages"
	"aibattle/pages/auth"
	"aibattle/pages/battle"
	"aibattle/pages/index"
	"aibattle/pages/middleware"
	"aibattle/pages/prompt"
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

			se.Router.GET("/{$}", index.Index(app, templ))

			withAuth(
				se.Router.GET("/logout", auth.Logout(app, templ)),
				se.Router.GET("/prompt", prompt.NewPromptForm(app, templ)),
				se.Router.POST("/prompt", prompt.CreatePrompt(app, templ)),
				se.Router.GET("/prompt/{id}", prompt.DetailedPrompt(app, templ)),
				se.Router.POST("/prompt/{id}", prompt.UpdatePrompt(app, templ)),
				se.Router.POST("/prompt/{id}/activate", prompt.ActivatePrompt(app, templ)),
				se.Router.GET("/battle", battle.List(app, templ)),
				se.Router.GET("/battle/{id}", battle.Detailed(app, templ)),
			)

			go func() {
				battler.RunBattleTask(app)
			}()

			go func() {
				ProcessPromts(app)
			}()
			return se.Next()
		},
	)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func ProcessPromts(app *pocketbase.PocketBase) {
	for {
		records, err := app.FindRecordsByFilter(
			"prompt",
			"status = ''",
			"-created",
			1,
			0,
		)

		if err != nil {
			log.Fatalf("Error fetching prompt: %v", err)
		}

		if len(records) > 0 {
			firstPrompt := records[0]
			newProg, promptErr := prompt.GetProgram(
				context.Background(), firstPrompt.Id, firstPrompt.GetString("text"),
			)
			if promptErr != nil {
				log.Printf("Error getting prompt: %v", promptErr)
				firstPrompt.Set("status", "error: "+promptErr.Error())
			}
			firstPrompt.Set("output", newProg)
			firstPrompt.Set("status", "done")
			saveErr := app.Save(firstPrompt)
			if saveErr != nil {
				log.Printf("Error saving prompt: %v", saveErr)
			}
		}

		time.Sleep(1 * time.Second)
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
