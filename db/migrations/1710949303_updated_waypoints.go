package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update
		edit_photo := &core.FileField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "tfhs3juh",
			"name": "photo",
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
					"image/svg+xml"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photo)
		collection.Fields.Add(edit_photo)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update
		edit_photo := &core.FileField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "tfhs3juh",
			"name": "photo",
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
					"image/svg+xml"
				],
				"thumbs": [],
				"maxSelect": 1,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photo)
		collection.Fields.Add(edit_photo)

		return app.Save(collection)
	})
}
