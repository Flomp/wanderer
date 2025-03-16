package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update
		edit_type := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "b57prsbu",
			"name": "type",
			"type": "select",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"trail_create",
					"list_create",
					"new_follower",
					"trail_comment",
					"trail_share",
					"list_share"
				]
			}
		}`), edit_type); err != nil {
			return err
		}
		collection.Fields.Add(edit_type)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update
		edit_type := &core.SelectField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "b57prsbu",
			"name": "type",
			"type": "select",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"trail_create",
					"t list_create",
					"new_follower",
					"trail_comment",
					"trail_share",
					"list_share"
				]
			}
		}`), edit_type); err != nil {
			return err
		}
		collection.Fields.Add(edit_type)

		return app.Save(collection)
	})
}
