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

		// add
		new_elevation_loss := &core.NumberField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "xutbwpq4",
			"name": "elevation_loss",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_elevation_loss)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("xutbwpq4")

		return app.Save(collection)
	})
}
