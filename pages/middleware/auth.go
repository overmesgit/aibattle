package middleware

import (
	"errors"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"net/http"
)

func LoadAuthToken() *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Id:       "cookie loader",
		Priority: 0,
		Func: func(e *core.RequestEvent) error {
			if e.Auth != nil {
				return e.Next()
			}

			token, err := e.Request.Cookie("token")
			if errors.As(err, &http.ErrNoCookie) {
				return e.Next()
			}
			if err != nil {
				return err
			}
			if token.Value == "" {
				return e.Next()
			}
			record, err := e.App.FindAuthRecordByToken(token.Value, core.TokenTypeAuth)
			if err != nil {
				e.App.Logger().Debug("loadAuthToken failure", "error", err)
			} else if record != nil {
				e.App.Logger().Debug("loadAuthToken success", "record", record.Id)
				e.Auth = record
			}

			return e.Next()
		},
	}
}
