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

		collection, err := dao.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// update
		edit_language := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "0sepzvkh",
			"name": "language",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"en",
					"de",
					"fr",
					"hu",
					"it",
					"nl",
					"pl",
					"pt",
					"zh"
				]
			}
		}`), edit_language); err != nil {
			return err
		}
		collection.Schema.AddField(edit_language)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("uavt73rsqcn1n13")
		if err != nil {
			return err
		}

		// update
		edit_language := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "0sepzvkh",
			"name": "language",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"en",
					"de",
					"fr",
					"hu",
					"nl",
					"pl",
					"pt",
					"zh"
				]
			}
		}`), edit_language); err != nil {
			return err
		}
		collection.Schema.AddField(edit_language)

		return dao.SaveCollection(collection)
	})
}
