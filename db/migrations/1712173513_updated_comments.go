package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		// add
		new_trail := &core.RelationField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "snrlpxar",
			"name": "trail",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "e864strfxo14pm4",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_trail)
		collection.Fields.Add(new_trail)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("snrlpxar")

		return app.Save(collection)
	})
}
