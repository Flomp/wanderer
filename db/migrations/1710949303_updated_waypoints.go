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

		collection, err := dao.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update
		edit_photo := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "tfhs3juh",
			"name": "photo",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/png",
					"image/vnd.mozilla.apng",
					"image/webp",
					"image/svg+xml"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photo)
		collection.Schema.AddField(edit_photo)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("goeo2ubp103rzp9")
		if err != nil {
			return err
		}

		// update
		edit_photo := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "tfhs3juh",
			"name": "photo",
			"type": "file",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"mimeTypes": [
					"image/jpeg",
					"image/png",
					"image/vnd.mozilla.apng",
					"image/webp",
					"image/svg+xml"
				],
				"thumbs": [],
				"maxSelect": 1,
				"maxSize": 5242880,
				"protected": false
			}
		}`), edit_photo)
		collection.Schema.AddField(edit_photo)

		return dao.SaveCollection(collection)
	})
}
