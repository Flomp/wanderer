package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "list.author = @request.auth.id || actor.user = @request.auth.id",
			"viewRule": "list.author = @request.auth.id || actor.user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation1148540665",
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
		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"listRule": "list.author = @request.auth.id || user = @request.auth.id",
			"viewRule": "list.author = @request.auth.id || user = @request.auth.id"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation1148540665")

		return app.Save(collection)
	})
}
