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

		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM users JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("nqiuah7b")

		// remove
		collection.Fields.RemoveById("rpzvilry")

		// remove
		collection.Fields.RemoveById("pfppgcfx")

		// remove
		collection.Fields.RemoveById("jrqp7c8d")

		// remove
		collection.Fields.RemoveById("xjcpg4su")

		// remove
		collection.Fields.RemoveById("cuj9awlq")

		// add
		new_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "265lcwyn",
			"name": "max_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_distance); err != nil {
			return err
		}
		collection.Fields.Add(new_max_distance)

		// add
		new_max_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "8udbn3hb",
			"name": "max_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(new_max_elevation_gain)

		// add
		new_max_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "0mw2f9gg",
			"name": "max_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_duration); err != nil {
			return err
		}
		collection.Fields.Add(new_max_duration)

		// add
		new_min_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "qumwuc7g",
			"name": "min_distance",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_distance); err != nil {
			return err
		}
		collection.Fields.Add(new_min_distance)

		// add
		new_min_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "sins29kl",
			"name": "min_elevation_gain",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(new_min_elevation_gain)

		// add
		new_min_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "dgfnqloa",
			"name": "min_duration",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_duration); err != nil {
			return err
		}
		collection.Fields.Add(new_min_duration)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT users.id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM users JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// add
		del_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_max_distance); err != nil {
			return err
		}
		collection.Fields.Add(del_max_distance)

		// add
		del_max_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_max_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(del_max_elevation_gain)

		// add
		del_max_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_max_duration); err != nil {
			return err
		}
		collection.Fields.Add(del_max_duration)

		// add
		del_min_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_min_distance); err != nil {
			return err
		}
		collection.Fields.Add(del_min_distance)

		// add
		del_min_elevation_gain := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_min_elevation_gain); err != nil {
			return err
		}
		collection.Fields.Add(del_min_elevation_gain)

		// add
		del_min_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
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
		}`), del_min_duration); err != nil {
			return err
		}
		collection.Fields.Add(del_min_duration)

		// remove
		collection.Fields.RemoveById("265lcwyn")

		// remove
		collection.Fields.RemoveById("8udbn3hb")

		// remove
		collection.Fields.RemoveById("0mw2f9gg")

		// remove
		collection.Fields.RemoveById("qumwuc7g")

		// remove
		collection.Fields.RemoveById("sins29kl")

		// remove
		collection.Fields.RemoveById("dgfnqloa")

		return app.Save(collection)
	})
}
