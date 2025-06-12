package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1995454416")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": "trail.author.user = @request.auth.id || trail.public = true || trail.trail_share_via_trail.actor.user ?= @request.auth.id || actor.user = @request.auth.id",
			"deleteRule": "actor.user = @request.auth.id",
			"listRule": "trail.author.user = @request.auth.id || trail.public = true || trail.trail_share_via_trail.actor.user ?= @request.auth.id || actor.user = @request.auth.id",
			"viewRule": "trail.author.user = @request.auth.id || trail.public = true || trail.trail_share_via_trail.actor.user ?= @request.auth.id || actor.user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1995454416")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"createRule": null,
			"deleteRule": null,
			"listRule": null,
			"viewRule": null
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
