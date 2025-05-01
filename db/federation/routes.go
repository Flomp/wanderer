package federation

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func Setup(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/.well-known/webfinger", webfinger)
		se.Router.GET("/activitypub/user/{username}", actor)
		se.Router.GET("/activitypub/user/{username}/outbox", get_outbox)
		se.Router.POST("/activitypub/user/{username}/outbox", post_outbox)

		return se.Next()
	})
}
