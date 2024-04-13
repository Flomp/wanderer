package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewRule = types.Pointer("@request.auth.id = id")

		// remove
		collection.Schema.RemoveField("iyhsoisl")

		// remove
		collection.Schema.RemoveField("kx2qfztr")

		// remove
		collection.Schema.RemoveField("z4qsnjeb")

		// remove
		collection.Schema.RemoveField("p66xomdb")

		// add
		new_max_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "osfey1yx",
			"name": "max_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_lat); err != nil {
			return err
		}
		collection.Schema.AddField(new_max_lat)

		// add
		new_max_lon := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "eohslzky",
			"name": "max_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_lon); err != nil {
			return err
		}
		collection.Schema.AddField(new_max_lon)

		// add
		new_min_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kigvankd",
			"name": "min_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_lat); err != nil {
			return err
		}
		collection.Schema.AddField(new_min_lat)

		// add
		new_min_lon := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ifnks9mg",
			"name": "min_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_lon); err != nil {
			return err
		}
		collection.Schema.AddField(new_min_lon)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewRule = nil

		// add
		del_max_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "iyhsoisl",
			"name": "max_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_lat); err != nil {
			return err
		}
		collection.Schema.AddField(del_max_lat)

		// add
		del_max_lon := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "kx2qfztr",
			"name": "max_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_lon); err != nil {
			return err
		}
		collection.Schema.AddField(del_max_lon)

		// add
		del_min_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "z4qsnjeb",
			"name": "min_lat",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_lat); err != nil {
			return err
		}
		collection.Schema.AddField(del_min_lat)

		// add
		del_min_lon := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "p66xomdb",
			"name": "min_lon",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_lon); err != nil {
			return err
		}
		collection.Schema.AddField(del_min_lon)

		// remove
		collection.Schema.RemoveField("osfey1yx")

		// remove
		collection.Schema.RemoveField("eohslzky")

		// remove
		collection.Schema.RemoveField("kigvankd")

		// remove
		collection.Schema.RemoveField("ifnks9mg")

		return dao.SaveCollection(collection)
	})
}
