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

		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance, COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, COALESCE(MAX(trails.duration), 0) AS max_duration, COALESCE(MIN(trails.distance), 0) AS min_distance, COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, COALESCE(MIN(trails.duration), 0) AS min_duration FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("jahzab8l")

		// remove
		collection.Fields.RemoveById("hz8wiojq")

		// remove
		collection.Fields.RemoveById("mhdukuxo")

		// remove
		collection.Fields.RemoveById("yl7yh4fi")

		// remove
		collection.Fields.RemoveById("mmdluwag")

		// remove
		collection.Fields.RemoveById("rruyzzld")

		// add
		new_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "n42yecz5",
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
			"id": "de8wl3e5",
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
			"id": "lqhk4k6e",
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
			"id": "ug5lh6ku",
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
			"id": "lkrr7zr4",
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
			"id": "sxhh02md",
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

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
		}`), &options); err != nil {
			return err
		}
		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// add
		del_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "jahzab8l",
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
			"id": "hz8wiojq",
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
			"id": "mhdukuxo",
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
			"id": "yl7yh4fi",
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
			"id": "mmdluwag",
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
			"id": "rruyzzld",
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
		collection.Fields.RemoveById("n42yecz5")

		// remove
		collection.Fields.RemoveById("de8wl3e5")

		// remove
		collection.Fields.RemoveById("lqhk4k6e")

		// remove
		collection.Fields.RemoveById("ug5lh6ku")

		// remove
		collection.Fields.RemoveById("lkrr7zr4")

		// remove
		collection.Fields.RemoveById("sxhh02md")

		return app.Save(collection)
	})
}
