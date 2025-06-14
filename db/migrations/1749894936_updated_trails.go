package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
			"hidden": false,
			"id": "aqbpyawe",
			"maxSelect": 99,
			"maxSize": 20971520,
			"mimeTypes": [
				"image/jpeg",
				"image/vnd.mozilla.apng",
				"image/png",
				"image/webp",
				"image/svg+xml",
				"image/heic",
				"video/mp4",
				"video/webm",
				"video/ogg"
			],
			"name": "photos",
			"presentable": false,
			"protected": false,
			"required": false,
			"system": false,
			"thumbs": [
				"600x0"
			],
			"type": "file"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
			"hidden": false,
			"id": "aqbpyawe",
			"maxSelect": 99,
			"maxSize": 20971520,
			"mimeTypes": [
				"image/jpeg",
				"image/vnd.mozilla.apng",
				"image/png",
				"image/webp",
				"image/svg+xml",
				"image/heic",
				"video/mp4",
				"video/webm",
				"video/ogg"
			],
			"name": "photos",
			"presentable": false,
			"protected": false,
			"required": false,
			"system": false,
			"thumbs": [],
			"type": "file"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
