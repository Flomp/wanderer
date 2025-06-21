package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "in1traur",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "old_follower",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "wxwomfd5",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "old_followee",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
                "hidden": false,
                "id": "select2063623452",
                "maxSelect": 1,
                "name": "status",
                "presentable": false,
                "required": true,
                "system": false,
                "type": "select",
                "values": [
                    "pending",
                    "accepted"
                ]
            }`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation3117812038",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "follower",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "pbc_1295301207",
			"hidden": false,
			"id": "relation973442177",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "followee",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "in1traur",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "follower",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "wxwomfd5",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "followee",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("relation3117812038")

		// remove field
		collection.Fields.RemoveById("relation973442177")

		return app.Save(collection)
	})
}
