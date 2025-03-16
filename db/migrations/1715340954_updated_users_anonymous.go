package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT id, username, avatar FROM users"

		// remove
		collection.Fields.RemoveById("qzkmy7nv")

		// add
		new_username := &core.TextField{}
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
		collection.Fields.Add(new_username)

		// add
		new_avatar := &core.FileField{}
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
		collection.Fields.Add(new_avatar)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT id, username FROM users"

		// add
		del_username := &core.TextField{}
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
		collection.Fields.Add(del_username)

		// remove
		collection.Fields.RemoveById("mbokwbb3")

		// remove
		collection.Fields.RemoveById("60wozbto")

		return app.Save(collection)
	})
}
