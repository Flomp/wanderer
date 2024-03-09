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

		// remove
		collection.Schema.RemoveField("mcqce8l9")

		// add
		new_thumbnail := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "k2giqyjq",
			"name": "thumbnail",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_thumbnail)
		collection.Schema.AddField(new_thumbnail)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("e864strfxo14pm4")
		if err != nil {
			return err
		}

		// add
		del_thumbnail := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "mcqce8l9",
			"name": "thumbnail",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_thumbnail)
		collection.Schema.AddField(del_thumbnail)

		// remove
		collection.Schema.RemoveField("k2giqyjq")

		return dao.SaveCollection(collection)
	})
}
