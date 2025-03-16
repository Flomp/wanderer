package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// add
		new_category := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "owlyzl1x",
			"name": "category",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "kjxvi8asj2igqwf",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_category); err != nil {
			return err
		}
		collection.Fields.Add(new_category)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// remove
		collection.Fields.RemoveById("owlyzl1x")

		return app.Save(collection)
	})
}
