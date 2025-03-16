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

		collection.ViewQuery = "SELECT id, username, avatar, created FROM users"

		// remove
		collection.Fields.RemoveById("ogrxf4ps")

		// remove
		collection.Fields.RemoveById("7k4bozux")

		// add
		new_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kvteagv6",
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
			"id": "m9efnpak",
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

		collection.ViewQuery = "SELECT id, username, avatar FROM users"

		// add
		del_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ogrxf4ps",
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

		// add
		del_avatar := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "7k4bozux",
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
		}`), del_avatar); err != nil {
			return err
		}
		collection.Fields.Add(del_avatar)

		// remove
		collection.Fields.RemoveById("kvteagv6")

		// remove
		collection.Fields.RemoveById("m9efnpak")

		return app.Save(collection)
	})
}
