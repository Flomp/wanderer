package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update
		edit_type := &schema.SchemaField{}
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
		collection.Schema.AddField(edit_type)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		// update
		edit_type := &schema.SchemaField{}
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
		collection.Schema.AddField(edit_type)

		return dao.SaveCollection(collection)
	})
}
