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

		// update
		edit_photos := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ixnksbkt",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/png",
					"image/vnd.mozilla.apng",
					"image/webp",
					"image/svg+xml",
					"image/heic"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 20971520,
				"protected": false
			}
		}`), edit_photos); err != nil {
			return err
		}
		collection.Fields.Add(edit_photos)

		// update
		edit_author := &core.RelationField{}
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
		}`), edit_author); err != nil {
			return err
		}
		collection.Fields.Add(edit_author)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// update
		edit_photos := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ixnksbkt",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/png",
					"image/vnd.mozilla.apng",
					"image/webp",
					"image/svg+xml",
					"image/heic"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photos); err != nil {
			return err
		}
		collection.Fields.Add(edit_photos)

		// update
		edit_author := &core.RelationField{}
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
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_author); err != nil {
			return err
		}
		collection.Fields.Add(edit_author)

		return app.Save(collection)
	})
}
