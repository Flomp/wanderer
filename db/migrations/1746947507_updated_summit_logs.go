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
			"deleteRule": "@request.auth.id != \"\" && (trails_via_summit_logs.author.user = @request.auth.id || author.user = @request.auth.id)",
			"listRule": "author.user = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)",
			"updateRule": "@request.auth.id != \"\" && (trails_via_summit_logs.author.user = @request.auth.id || author.user = @request.auth.id)",
			"viewRule": "author.user = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("r0mj3tkr")

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation3182418120",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "author",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
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
			"deleteRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author.user ?= @request.auth.id) || (author = @request.auth.id)",
			"listRule": "author = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)",
			"updateRule": "@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author.user ?= @request.auth.id) || (author = @request.auth.id)",
			"viewRule": "author = @request.auth.id || trails_via_summit_logs.author.user ?= @request.auth.id || trails_via_summit_logs.public ?= true || \n(@collection.trail_share.trail.id ?= trails_via_summit_logs.id && @collection.trail_share.user ?= @request.auth.id)"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "r0mj3tkr",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "author",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation3182418120")

		return app.Save(collection)
	})
}
