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

		// remove
		collection.Fields.RemoveById("mcqce8l9")

		// add
		new_thumbnail := &core.NumberField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "k2giqyjq",
			"name": "thumbnail",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_thumbnail)
		collection.Fields.Add(new_thumbnail)

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// add
		del_thumbnail := &core.TextField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "mcqce8l9",
			"name": "thumbnail",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_thumbnail)
		collection.Fields.Add(del_thumbnail)

		// remove
		collection.Fields.RemoveById("k2giqyjq")

		return app.Save(collection)
	})
}
