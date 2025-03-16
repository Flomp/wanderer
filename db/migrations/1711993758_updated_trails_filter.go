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

		collection.ViewQuery = "SELECT (ROW_NUMBER() OVER()) as id, MAX(trails.distance) AS max_distance, MAX(trails.elevation_gain) AS max_elevation_gain, MAX(trails.duration) AS max_duration, MIN(trails.distance) AS min_distance, MIN(trails.elevation_gain) AS min_elevation_gain, MIN(trails.duration) AS min_duration FROM trails;"

		// remove
		collection.Fields.RemoveById("ovledory")

		// add
		new_max_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "3neuurse",
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
			"id": "oynpksx3",
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
			"id": "tkoppkjw",
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
			"id": "uxc0eimu",
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
			"id": "qcktbht6",
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
			"id": "qpkhyujt",
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

		collection.ViewQuery = "SELECT (ROW_NUMBER() OVER()) as id, MAX(trails.distance) AS max_distance FROM trails;"

		// add
		del_max_distance := &core.JSONField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ovledory",
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

		// remove
		collection.Fields.RemoveById("3neuurse")

		// remove
		collection.Fields.RemoveById("oynpksx3")

		// remove
		collection.Fields.RemoveById("tkoppkjw")

		// remove
		collection.Fields.RemoveById("uxc0eimu")

		// remove
		collection.Fields.RemoveById("qcktbht6")

		// remove
		collection.Fields.RemoveById("qpkhyujt")

		return app.Save(collection)
	})
}
