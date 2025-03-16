package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT users.id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM users JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("54a1hvuq")

		// remove
		collection.Fields.RemoveById("stcnunkf")

		// remove
		collection.Fields.RemoveById("bholagx7")

		// remove
		collection.Fields.RemoveById("qjrtggwi")

		// remove
		collection.Fields.RemoveById("wtlhqwd2")

		// remove
		collection.Fields.RemoveById("sputfuyq")

		// add
		new_max_distance := &core.JSONField{}
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
		collection.Fields.Add(new_max_distance)

		// add
		new_max_elevation_gain := &core.JSONField{}
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
		collection.Fields.Add(new_max_elevation_gain)

		// add
		new_max_duration := &core.JSONField{}
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
		collection.Fields.Add(new_max_duration)

		// add
		new_min_distance := &core.JSONField{}
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
		collection.Fields.Add(new_min_distance)

		// add
		new_min_elevation_gain := &core.JSONField{}
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
		collection.Fields.Add(new_min_elevation_gain)

		// add
		new_min_duration := &core.JSONField{}
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
		collection.Fields.Add(new_min_duration)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT (ROW_NUMBER() OVER()) as id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM trails;"

		// add
		del_max_distance := &core.JSONField{}
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
		collection.Fields.Add(del_max_distance)

		// add
		del_max_elevation_gain := &core.JSONField{}
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
		collection.Fields.Add(del_max_elevation_gain)

		// add
		del_max_duration := &core.JSONField{}
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
		collection.Fields.Add(del_max_duration)

		// add
		del_min_distance := &core.JSONField{}
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
		collection.Fields.Add(del_min_distance)

		// add
		del_min_elevation_gain := &core.JSONField{}
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
		collection.Fields.Add(del_min_elevation_gain)

		// add
		del_min_duration := &core.JSONField{}
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
		collection.Fields.Add(del_min_duration)

		// remove
		collection.Fields.RemoveById("5udo4avx")

		// remove
		collection.Fields.RemoveById("omuxiatj")

		// remove
		collection.Fields.RemoveById("yq4snqnf")

		// remove
		collection.Fields.RemoveById("q8d41loj")

		// remove
		collection.Fields.RemoveById("xsetw0o6")

		// remove
		collection.Fields.RemoveById("n4pq80mu")

		return app.Save(collection)
	})
}
