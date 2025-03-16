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
		new_date := &core.DateField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "hovyvbtt",
			"name": "date",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), new_date); err != nil {
			return err
		}
		collection.Fields.Add(new_date)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("hovyvbtt")

		return app.Save(collection)
	})
}
