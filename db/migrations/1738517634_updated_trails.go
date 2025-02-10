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

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// add
		new_external_id := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "sajmiuau",
			"name": "external_id",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_external_id); err != nil {
			return err
		}
		collection.Schema.AddField(new_external_id)

		// add
		new_external_provider := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "htr35nha",
			"name": "external_provider",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"strava"
				]
			}
		}`), new_external_provider); err != nil {
			return err
		}
		collection.Schema.AddField(new_external_provider)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("sajmiuau")

		// remove
		collection.Schema.RemoveField("htr35nha")

		return dao.SaveCollection(collection)
	})
}
