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

		collection, err := dao.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "SELECT users.id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM users JOIN trails ON users.id = trails.author GROUP BY users.id;"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("54a1hvuq")

		// remove
		collection.Schema.RemoveField("stcnunkf")

		// remove
		collection.Schema.RemoveField("bholagx7")

		// remove
		collection.Schema.RemoveField("qjrtggwi")

		// remove
		collection.Schema.RemoveField("wtlhqwd2")

		// remove
		collection.Schema.RemoveField("sputfuyq")

		// add
		new_max_distance := &schema.SchemaField{}
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
		}`), new_max_distance)
		collection.Schema.AddField(new_max_distance)

		// add
		new_max_elevation_gain := &schema.SchemaField{}
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
		}`), new_max_elevation_gain)
		collection.Schema.AddField(new_max_elevation_gain)

		// add
		new_max_duration := &schema.SchemaField{}
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
		}`), new_max_duration)
		collection.Schema.AddField(new_max_duration)

		// add
		new_min_distance := &schema.SchemaField{}
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
		}`), new_min_distance)
		collection.Schema.AddField(new_min_distance)

		// add
		new_min_elevation_gain := &schema.SchemaField{}
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
		}`), new_min_elevation_gain)
		collection.Schema.AddField(new_min_elevation_gain)

		// add
		new_min_duration := &schema.SchemaField{}
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
		}`), new_min_duration)
		collection.Schema.AddField(new_min_duration)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "SELECT (ROW_NUMBER() OVER()) as id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM trails;"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_max_distance := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "54a1hvuq",
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
			"id": "stcnunkf",
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
			"id": "bholagx7",
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
			"id": "qjrtggwi",
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
			"id": "wtlhqwd2",
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
			"id": "sputfuyq",
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

		return dao.SaveCollection(collection)
	})
}
