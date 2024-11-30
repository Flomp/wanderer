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

		collection, err := dao.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// add
		new_photos := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ixnksbkt",
			"name": "photos",
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
					"image/svg+xml",
					"image/heic"
				],
				"thumbs": [],
				"maxSelect": 99,
				"maxSize": 5242880,
				"protected": false
			}
		}`), new_photos); err != nil {
			return err
		}
		collection.Schema.AddField(new_photos)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("dd2l9a4vxpy2ni8")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("ixnksbkt")

		return dao.SaveCollection(collection)
	})
}
