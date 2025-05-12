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
			"deleteRule": "@request.auth.id != \"\" && (trail.author.user = @request.auth.id || author.user = @request.auth.id)",
			"listRule": "author.user = @request.auth.id || trail.author.user ?= @request.auth.id || trail.public ?= true || \ntrail.trail_share_via_trail.user ?= @request.auth.id",
			"updateRule": "@request.auth.id != \"\" && (trail.author.user = @request.auth.id || author.user = @request.auth.id)",
			"viewRule": "author.user = @request.auth.id || trail.author.user ?= @request.auth.id || trail.public ?= true || \ntrail.trail_share_via_trail.user ?= @request.auth.id"
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
			"deleteRule": "@request.auth.id != \"\" && (trails_via_summit_logs.author.user = @request.auth.id || author.user = @request.auth.id)",
			"listRule": "author.user = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)",
			"updateRule": "@request.auth.id != \"\" && (trails_via_summit_logs.author.user = @request.auth.id || author.user = @request.auth.id)",
			"viewRule": "author.user = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
