package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update
		edit_photos := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "aqbpyawe",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/vnd.mozilla.apng",
					"image/png",
					"image/webp",
					"image/svg+xml",
					"image/heic",
					"video/mp4"
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

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update
		edit_photos := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "aqbpyawe",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/vnd.mozilla.apng",
					"image/png",
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

		return app.Save(collection)
	})
}
