package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "(@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false\n&&\n@collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )) && @request.auth.id = author",
			"viewRule": "(@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false\n&&\n@collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )) && @request.auth.id = author"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("t9lphichi5xwyeu")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)",
			"viewRule": "(\n    @request.auth.id = author || \n    (@collection.users_anonymous.id ?= author && @collection.users_anonymous.private ?= false)\n)\n&&\n(\n    @collection.trails.id ?= trail_id && \n    (\n        @collection.trails.author ?= @request.auth.id ||\n        @collection.trails.public ?= true || \n        (@request.auth.id != \"\" && @collection.trails.trail_share_via_trail.user ?= @request.auth.id)\n    )\n)"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
