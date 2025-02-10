package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || public = true || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		// add
		new_public := &schema.SchemaField{}
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
		collection.Schema.AddField(new_public)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		collection.ViewRule = types.Pointer("author = @request.auth.id || (@request.auth.id != \"\" && list_share_via_list.user ?= @request.auth.id)")

		// remove
		collection.Schema.RemoveField("rolk3q3j")

		return dao.SaveCollection(collection)
	})
}
