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
		new_difficulty := &core.SelectField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "dywtnynw",
			"name": "difficulty",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"easy",
					"moderate",
					"difficult"
				]
			}
		}`), new_difficulty)
		collection.Fields.Add(new_difficulty)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("dywtnynw")

		return app.Save(collection)
	})
}
