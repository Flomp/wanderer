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

		// update
		edit_user := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mix12kkh",
			"name": "user",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_user); err != nil {
			return err
		}
		collection.Fields.Add(edit_user)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update
		edit_user := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "mix12kkh",
			"name": "user",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_user); err != nil {
			return err
		}
		collection.Fields.Add(edit_user)

		return app.Save(collection)
	})
}
