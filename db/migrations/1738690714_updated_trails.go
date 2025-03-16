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

		// update
		edit_external_provider := &core.SelectField{}
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
					"strava",
					"komoot"
				]
			}
		}`), edit_external_provider); err != nil {
			return err
		}
		collection.Fields.Add(edit_external_provider)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update
		edit_external_provider := &core.SelectField{}
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
		}`), edit_external_provider); err != nil {
			return err
		}
		collection.Fields.Add(edit_external_provider)

		return app.Save(collection)
	})
}
