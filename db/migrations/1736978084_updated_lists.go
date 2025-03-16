package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		// update
		edit_author := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kwm6zdet",
			"name": "author",
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
		}`), edit_author); err != nil {
			return err
		}
		collection.Fields.Add(edit_author)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("r6gu2ajyidy1x69")
		if err != nil {
			return err
		}

		// update
		edit_author := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kwm6zdet",
			"name": "author",
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
		}`), edit_author); err != nil {
			return err
		}
		collection.Fields.Add(edit_author)

		return app.Save(collection)
	})
}
