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
			"listRule": "author = @request.auth.id || trails_via_summit_logs.author ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)",
			"viewRule": "author = @request.auth.id || trails_via_summit_logs.author ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)"
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
			"viewRule": "(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || \n(@collection.trails.summit_logs.id ?= id && @collection.trails.public ?= true) || \n(@collection.trail_share.trail.summit_logs.id ?= id && @collection.trail_share.user ?= @request.auth.id) ||\n(author = @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
