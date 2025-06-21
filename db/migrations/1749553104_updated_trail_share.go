package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation2375276105",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "actor",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1mns8mlal6uf9ku")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation2375276105")

		return app.UnsafeWithoutHooks().Save(collection)
	})
}
