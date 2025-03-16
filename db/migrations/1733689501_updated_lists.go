package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		// add
		new_public := &core.BoolField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "rolk3q3j",
			"name": "public",
			"type": "bool",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_public); err != nil {
			return err
		}
		collection.Fields.Add(new_public)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		// remove
		collection.Fields.RemoveById("rolk3q3j")

		return app.Save(collection)
	})
}
