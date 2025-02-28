package main

import (
	"aibattle/battler"
	_ "aibattle/migrations"
	"aibattle/pages"
	"aibattle/pages/auth"
	"aibattle/pages/battle"
	"aibattle/pages/builder"
	"aibattle/pages/index"
	"aibattle/pages/leader"
	"aibattle/pages/middleware"
	"aibattle/pages/prompt"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/dbx"
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
				se.Router.POST("/prompt/{id}/activate", prompt.ActivatePrompt(app, templ)),
				se.Router.GET("/battle", battle.List(app, templ)),
				se.Router.GET("/battle/{id}", battle.Detailed(app, templ)),
			)

			go func() {
				if os.Getenv("DISABLE_BATTLE") != "true" {
					battler.RunBattleTask(app)
				}
			}()

			go func() {
				ProcessPrompts(app)
			}()
			return se.Next()
		},
	)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func ProcessPrompts(app *pocketbase.PocketBase) {
	ScheduleRemainingPrompts(app)
	for {
		nextPrompt := <-prompt.PromptsToProcess
		newProg, promptErr := builder.GetProgram(
			context.Background(), nextPrompt.Id, nextPrompt.GetString("text"),
			nextPrompt.GetString("language"),
		)
		if promptErr != nil {
			log.Printf("Error getting prompt: %v", promptErr)
			nextPrompt.Set("status", "error")
			nextPrompt.Set("error", promptErr.Error())
		} else {
			nextPrompt.Set("status", "done")
			nextPrompt.Set("error", "")
		}
		nextPrompt.Set("output", newProg)
		saveErr := app.Save(nextPrompt)
		if saveErr != nil {
			log.Printf("Error saving prompt: %v", saveErr)
		}
		activateIfFirstPrompt(app, nextPrompt)
	}
}

func activateIfFirstPrompt(app *pocketbase.PocketBase, prompt *core.Record) error {
	if prompt.GetBool("active") {
		return nil
	}
	user := prompt.GetString("user")
	records, err := app.FindRecordsByFilter(
		"prompt",
		"user = {:user} && active = true",
		"-created",
		1,
		0,
		dbx.Params{
			"user": user,
		},
	)
	if err != nil {
		log.Printf("Error checking active prompts: %v", err)
		return err
	}
	if len(records) == 0 {
		prompt.Set("active", true)
		saveError := app.Save(prompt)
		if saveError != nil {
			log.Printf("Error activating prompt: %v", saveError)
			return saveError
		}

		battler.BattleChannel <- prompt.Id
	}
	return nil
}

func ScheduleRemainingPrompts(app *pocketbase.PocketBase) {
	records, err := app.FindRecordsByFilter(
		"prompt",
		"status = ''",
		"-created",
		20,
		0,
	)
	if err != nil {
		log.Fatalf("Error fetching prompt: %v", err)
	}
	for _, record := range records {
		log.Printf("Scheduling record: %v", record.Id)
		prompt.PromptsToProcess <- record
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
