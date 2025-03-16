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
		new_external_id := &core.TextField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "sajmiuau",
			"name": "external_id",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_external_id); err != nil {
			return err
		}
		collection.Fields.Add(new_external_id)

		// add
		new_external_provider := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "htr35nha",
			"name": "external_provider",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"strava"
				]
			}
		}`), new_external_provider); err != nil {
			return err
		}
		collection.Fields.Add(new_external_provider)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("sajmiuau")

		// remove
		collection.Fields.RemoveById("htr35nha")

		return app.Save(collection)
	})
}
