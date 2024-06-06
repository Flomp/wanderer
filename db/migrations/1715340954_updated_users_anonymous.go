package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id, username, avatar FROM users"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("qzkmy7nv")

		// add
		new_username := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mbokwbb3",
			"name": "username",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_username); err != nil {
			return err
		}
		collection.Schema.AddField(new_username)

		// add
		new_avatar := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "60wozbto",
			"name": "avatar",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/png",
					"image/svg+xml",
					"image/gif",
					"image/webp"
				],
				"thumbs": null,
				"maxSelect": 1,
				"maxSize": 5242880,
				"protected": false
			}
		}`), new_avatar); err != nil {
			return err
		}
		collection.Schema.AddField(new_avatar)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT id, username FROM users"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// add
		del_username := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "qzkmy7nv",
			"name": "username",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_username); err != nil {
			return err
		}
		collection.Schema.AddField(del_username)

		// remove
		collection.Schema.RemoveField("mbokwbb3")

		// remove
		collection.Schema.RemoveField("60wozbto")

		return dao.SaveCollection(collection)
	})
}
