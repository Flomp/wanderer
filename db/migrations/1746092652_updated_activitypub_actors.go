package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1295301207")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text2115105593")

		// remove field
		collection.Fields.RemoveById("text1793578352")

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text3458754147",
			"max": 0,
			"min": 0,
			"name": "summary",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"hidden": false,
			"id": "date1748787223",
			"max": "",
			"min": "",
			"name": "published",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "date"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url2115105593",
			"name": "inbox",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url1793578352",
			"name": "outbox",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url2215181735",
			"name": "followers",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(10, []byte(`{
			"exceptDomains": null,
			"hidden": false,
			"id": "url1908379107",
			"name": "following",
			"onlyDomains": null,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "url"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1295301207")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text2115105593",
			"max": 0,
			"min": 0,
			"name": "inbox",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1793578352",
			"max": 0,
			"min": 0,
			"name": "outbox",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text3458754147")

		// remove field
		collection.Fields.RemoveById("date1748787223")

		// remove field
		collection.Fields.RemoveById("url2115105593")

		// remove field
		collection.Fields.RemoveById("url1793578352")

		// remove field
		collection.Fields.RemoveById("url2215181735")

		// remove field
		collection.Fields.RemoveById("url1908379107")

		return app.Save(collection)
	})
}
