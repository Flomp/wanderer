package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update
		edit_follower := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "in1traur",
			"name": "follower",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_follower); err != nil {
			return err
		}
		collection.Fields.Add(edit_follower)

		// update
		edit_followee := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "wxwomfd5",
			"name": "followee",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_followee); err != nil {
			return err
		}
		collection.Fields.Add(edit_followee)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		// update
		edit_follower := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "in1traur",
			"name": "follower",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_follower); err != nil {
			return err
		}
		collection.Fields.Add(edit_follower)

		// update
		edit_followee := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "wxwomfd5",
			"name": "followee",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_followee); err != nil {
			return err
		}
		collection.Fields.Add(edit_followee)

		return app.Save(collection)
	})
}
