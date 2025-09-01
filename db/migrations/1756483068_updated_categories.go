package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("kjxvi8asj2igqwf")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(23, []byte(`{
			"hidden": false,
			"id": "number957275141",
			"max": null,
			"min": 0,
			"name": "wp_merge_radius",
			"onlyInt": true,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("kjxvi8asj2igqwf")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("number957275141")

		return app.Save(collection)
	})
}
