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

		// add
		new_photos := &core.FileField{}
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
		}`), new_photos); err != nil {
			return err
		}
		collection.Fields.Add(new_photos)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("ixnksbkt")

		return app.Save(collection)
	})
}
