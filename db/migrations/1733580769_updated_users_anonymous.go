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

		collection.ViewQuery = "SELECT id, username, avatar, bio, created FROM users"

		// remove
		collection.Fields.RemoveById("femc0ok5")

		// remove
		collection.Fields.RemoveById("aglcodpn")

		// add
		new_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "tkxbpwxe",
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
			"id": "xs5qvkih",
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

		// add
		new_bio := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "agsa8h5z",
			"name": "bio",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": 10000,
				"pattern": ""
			}
		}`), new_bio); err != nil {
			return err
		}
		collection.Fields.Add(new_bio)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT id, username, avatar, created FROM users"

		// add
		del_username := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "femc0ok5",
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
			"id": "aglcodpn",
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
		collection.Fields.RemoveById("tkxbpwxe")

		// remove
		collection.Fields.RemoveById("xs5qvkih")

		// remove
		collection.Fields.RemoveById("agsa8h5z")

		return app.Save(collection)
	})
}
