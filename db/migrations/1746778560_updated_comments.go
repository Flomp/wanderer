package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("fhgxdiam")

		// remove field
		collection.Fields.RemoveById("7lwo1mxx")

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation3182418120",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "author",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "e864strfxo14pm4",
			"hidden": false,
			"id": "snrlpxar",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "trail",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("lf06qip3f4d11yk")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"hidden": false,
			"id": "fhgxdiam",
			"max": null,
			"min": null,
			"name": "rating",
			"onlyInt": false,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "7lwo1mxx",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "author",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation3182418120")

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"cascadeDelete": true,
			"collectionId": "e864strfxo14pm4",
			"hidden": false,
			"id": "snrlpxar",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "trail",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
