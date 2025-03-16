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
		edit_list := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "luqrtipy",
			"name": "list",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "r6gu2ajyidy1x69",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_list); err != nil {
			return err
		}
		collection.Fields.Add(edit_list)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("1kot7t9na3hi0gl")
		if err != nil {
			return err
		}

		// update
		edit_list := &core.RelationField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "luqrtipy",
			"name": "list",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "r6gu2ajyidy1x69",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_list); err != nil {
			return err
		}
		collection.Fields.Add(edit_list)

		return app.Save(collection)
	})
}
