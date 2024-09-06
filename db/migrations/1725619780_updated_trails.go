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

		// update
		edit_photos := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "aqbpyawe",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/vnd.mozilla.apng",
					"image/png",
					"image/webp",
					"image/svg+xml",
					"image/heic"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photos); err != nil {
			return err
		}
		collection.Schema.AddField(edit_photos)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// update
		edit_photos := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "aqbpyawe",
			"name": "photos",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/vnd.mozilla.apng",
					"image/png",
					"image/webp",
					"image/svg+xml"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photos); err != nil {
			return err
		}
		collection.Schema.AddField(edit_photos)

		return dao.SaveCollection(collection)
	})
}
