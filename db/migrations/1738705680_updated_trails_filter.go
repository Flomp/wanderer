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

		collection.ViewQuery = "SELECT users.id, COALESCE(printf(\"%.2f\", MAX(trails.distance)), 0) AS max_distance,\n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_gain)), 0) AS max_elevation_gain, \n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_loss)), 0) AS max_elevation_loss, \n  COALESCE(printf(\"%.2f\", MAX(trails.duration)), 0) AS max_duration, \n  COALESCE(printf(\"%.2f\", MIN(trails.distance)), 0) AS min_distance,   \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_gain)), 0) AS min_elevation_gain, \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_loss)), 0) AS min_elevation_loss, \n  COALESCE(printf(\"%.2f\", MIN(trails.duration)), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("h18nofwq")

		// remove
		collection.Fields.RemoveById("thhrslvh")

		// remove
		collection.Fields.RemoveById("e2bbiudy")

		// remove
		collection.Fields.RemoveById("smzzpwpp")

		// remove
		collection.Fields.RemoveById("1b2horvp")

		// remove
		collection.Fields.RemoveById("m8ghv383")

		// remove
		collection.Fields.RemoveById("awhc47b7")

		// remove
		collection.Fields.RemoveById("q2houpyd")

		// add
		new_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "sds7j8bk",
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
			"id": "olgpl62o",
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
		new_max_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "hsacrn8r",
			"name": "max_elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_max_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_max_elevation_loss)

		// add
		new_max_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ctoyvjxp",
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
			"id": "jlwabean",
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
			"id": "w3zecui8",
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
		new_min_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "nby5xyd9",
			"name": "min_elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), new_min_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(new_min_elevation_loss)

		// add
		new_min_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "rn0ije3h",
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

		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance,\n  COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, \n  COALESCE(MAX(trails.elevation_loss), 0) AS max_elevation_loss, \n  COALESCE(MAX(trails.duration), 0) AS max_duration, \n  COALESCE(MIN(trails.distance), 0) AS min_distance,   \n  COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, \n  COALESCE(MIN(trails.elevation_loss), 0) AS min_elevation_loss, \n  COALESCE(MIN(trails.duration), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"

		// add
		del_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "h18nofwq",
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
			"id": "thhrslvh",
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
		del_max_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "e2bbiudy",
			"name": "max_elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_max_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(del_max_elevation_loss)

		// add
		del_max_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "smzzpwpp",
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
			"id": "1b2horvp",
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
			"id": "m8ghv383",
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
		del_min_elevation_loss := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "awhc47b7",
			"name": "min_elevation_loss",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSize": 1
			}
		}`), del_min_elevation_loss); err != nil {
			return err
		}
		collection.Fields.Add(del_min_elevation_loss)

		// add
		del_min_duration := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "q2houpyd",
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
		collection.Fields.RemoveById("sds7j8bk")

		// remove
		collection.Fields.RemoveById("olgpl62o")

		// remove
		collection.Fields.RemoveById("hsacrn8r")

		// remove
		collection.Fields.RemoveById("ctoyvjxp")

		// remove
		collection.Fields.RemoveById("jlwabean")

		// remove
		collection.Fields.RemoveById("w3zecui8")

		// remove
		collection.Fields.RemoveById("nby5xyd9")

		// remove
		collection.Fields.RemoveById("rn0ije3h")

		return app.Save(collection)
	})
}
