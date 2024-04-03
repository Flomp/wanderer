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

		collection, err := dao.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ListRule = nil

		collection.ViewRule = types.Pointer("@request.auth.id = id")

		// remove
		collection.Schema.RemoveField("5udo4avx")

		// remove
		collection.Schema.RemoveField("omuxiatj")

		// remove
		collection.Schema.RemoveField("yq4snqnf")

		// remove
		collection.Schema.RemoveField("q8d41loj")

		// remove
		collection.Schema.RemoveField("xsetw0o6")

		// remove
		collection.Schema.RemoveField("n4pq80mu")

		// add
		new_max_distance := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "nqiuah7b",
			"name": "max_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_distance)
		collection.Schema.AddField(new_max_distance)

		// add
		new_max_elevation_gain := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "rpzvilry",
			"name": "max_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_elevation_gain)
		collection.Schema.AddField(new_max_elevation_gain)

		// add
		new_max_duration := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "pfppgcfx",
			"name": "max_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_duration)
		collection.Schema.AddField(new_max_duration)

		// add
		new_min_distance := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "jrqp7c8d",
			"name": "min_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_distance)
		collection.Schema.AddField(new_min_distance)

		// add
		new_min_elevation_gain := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xjcpg4su",
			"name": "min_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_elevation_gain)
		collection.Schema.AddField(new_min_elevation_gain)

		// add
		new_min_duration := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cuj9awlq",
			"name": "min_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_duration)
		collection.Schema.AddField(new_min_duration)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\"")

		collection.ViewRule = nil

		// add
		del_max_distance := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5udo4avx",
			"name": "max_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_distance)
		collection.Schema.AddField(del_max_distance)

		// add
		del_max_elevation_gain := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "omuxiatj",
			"name": "max_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_elevation_gain)
		collection.Schema.AddField(del_max_elevation_gain)

		// add
		del_max_duration := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "yq4snqnf",
			"name": "max_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_duration)
		collection.Schema.AddField(del_max_duration)

		// add
		del_min_distance := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "q8d41loj",
			"name": "min_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_distance)
		collection.Schema.AddField(del_min_distance)

		// add
		del_min_elevation_gain := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xsetw0o6",
			"name": "min_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_elevation_gain)
		collection.Schema.AddField(del_min_elevation_gain)

		// add
		del_min_duration := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "n4pq80mu",
			"name": "min_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_duration)
		collection.Schema.AddField(del_min_duration)

		// remove
		collection.Schema.RemoveField("nqiuah7b")

		// remove
		collection.Schema.RemoveField("rpzvilry")

		// remove
		collection.Schema.RemoveField("pfppgcfx")

		// remove
		collection.Schema.RemoveField("jrqp7c8d")

		// remove
		collection.Schema.RemoveField("xjcpg4su")

		// remove
		collection.Schema.RemoveField("cuj9awlq")

		return dao.SaveCollection(collection)
	})
}
