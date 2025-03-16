package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true) || (author = @request.auth.id)")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true) || (author = @request.auth.id)")

		collection.UpdateRule = types.Pointer("@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id) || (author = @request.auth.id)")

		collection.DeleteRule = types.Pointer("@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id) || (author = @request.auth.id)")

		// add
		new_author := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "r0mj3tkr",
			"name": "author",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_author); err != nil {
			return err
		}
		collection.Fields.Add(new_author)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true)")

		collection.ViewRule = types.Pointer("(@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && @collection.trails.author ?= @request.auth.id) || (@collection.trails.summit_logs.id ?= id && @collection.trails.public = true)")

		collection.UpdateRule = types.Pointer("@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id)")

		collection.DeleteRule = types.Pointer("@request.auth.id != \"\" && @collection.trails.summit_logs.id ?= id && (@collection.trails.author ?= @request.auth.id)")

		// remove
		collection.Fields.RemoveById("r0mj3tkr")

		return app.Save(collection)
	})
}
