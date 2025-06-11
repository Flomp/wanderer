package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "(@request.auth.id != \"\" && (author.user = @request.auth.id || trail.author.user = @request.auth.id)) || trail.public = true || trail.trail_share_via_trail.actor.user ?= @request.auth.id",
			"viewRule": "(@request.auth.id != \"\" && (author.user = @request.auth.id || trail.author.user = @request.auth.id)) || trail.public = true || trail.trail_share_via_trail.actor.user ?= @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "(@request.auth.id != \"\" && (author.user = @request.auth.id || trail.author.user = @request.auth.id)) || trail.public = true || (@collection.trail_share.trail ?= trail && @collection.trail_share.actor.user ?= @request.auth.id)",
			"viewRule": "(@request.auth.id != \"\" && (author.user = @request.auth.id || trail.author.user = @request.auth.id)) || trail.public = true || (@collection.trail_share.trail ?= trail && @collection.trail_share.actor.user ?= @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
