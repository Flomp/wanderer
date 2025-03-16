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
		new_gpx := &core.FileField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "rfwmdcpt",
			"name": "gpx",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [],
				"thumbs": [],
				"maxSelect": 1,
				"maxSize": 5242880,
				"protected": false
			}
		}`), new_gpx); err != nil {
			return err
		}
		collection.Fields.Add(new_gpx)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("rfwmdcpt")

		return app.Save(collection)
	})
}
