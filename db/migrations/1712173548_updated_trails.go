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
		collection.Fields.RemoveById("phhhroww")

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// add
		del_comments := &core.RelationField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "phhhroww",
			"name": "comments",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "lf06qip3f4d11yk",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": null
			}
		}`), del_comments)
		collection.Fields.Add(del_comments)

		return app.Save(collection)
	})
}
