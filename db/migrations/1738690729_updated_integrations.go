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

		collection, err := dao.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		// add
		new_komoot := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "6s0oxqgp",
			"name": "komoot",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 2000000
			}
		}`), new_komoot); err != nil {
			return err
		}
		collection.Schema.AddField(new_komoot)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("6s0oxqgp")

		return dao.SaveCollection(collection)
	})
}
