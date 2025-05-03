package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3752774184")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"exceptDomains": [],
			"hidden": false,
			"id": "url2434853685",
			"name": "iri",
			"onlyDomains": [],
			"presentable": false,
			"required": true,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text2363381545",
			"max": 0,
			"min": 0,
			"name": "type",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": true,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"hidden": false,
			"id": "json3616002756",
			"maxSize": 0,
			"name": "to",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "json"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"hidden": false,
			"id": "date1748787223",
			"max": "",
			"min": "",
			"name": "published",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "date"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"hidden": false,
			"id": "json2893285722",
			"maxSize": 0,
			"name": "object",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "json"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation3182418120",
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
		collection, err := app.FindCollectionByNameOrId("pbc_3752774184")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("url2434853685")

		// remove field
		collection.Fields.RemoveById("text2363381545")

		// remove field
		collection.Fields.RemoveById("json3616002756")

		// remove field
		collection.Fields.RemoveById("date1748787223")

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"hidden": false,
			"id": "json2893285722",
			"maxSize": 0,
			"name": "activity",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "json"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": false,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation3182418120",
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

		return app.Save(collection)
	})
}
