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

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// update
		edit_language := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t1wlsqyp",
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
		}`), edit_language)
		collection.Schema.AddField(edit_language)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// update
		edit_language := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t1wlsqyp",
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
					"pt"
				]
			}
		}`), edit_language)
		collection.Schema.AddField(edit_language)

		return dao.SaveCollection(collection)
	})
}
