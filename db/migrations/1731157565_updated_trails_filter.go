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

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
		}`), &options); err != nil {
			return err
		}
		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance,\n  COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, \n  COALESCE(MAX(trails.elevation_loss), 0) AS max_elevation_loss, \n  COALESCE(MAX(trails.duration), 0) AS max_duration, \n  COALESCE(MIN(trails.distance), 0) AS min_distance,   \n  COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, \n  COALESCE(MIN(trails.elevation_loss), 0) AS min_elevation_loss, \n  COALESCE(MIN(trails.duration), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("ixfa8u8m")

		// remove
		collection.Fields.RemoveById("bhn0jj40")

		// remove
		collection.Fields.RemoveById("3kujiwzh")

		// remove
		collection.Fields.RemoveById("yf3zwzn4")

		// remove
		collection.Fields.RemoveById("skvtkith")

		// remove
		collection.Fields.RemoveById("gydvo1ug")

		// add
		new_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "oyvx8oaq",
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
			"id": "4xu7voeh",
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
			"id": "sqzilovv",
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
			"id": "rupidtew",
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
			"id": "2y7dbhf9",
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
			"id": "ooiwexpc",
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
			"id": "ondmfwce",
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
			"id": "kcrhzmaw",
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
		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance, COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, COALESCE(MAX(trails.duration), 0) AS max_duration, COALESCE(MIN(trails.distance), 0) AS min_distance, COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, COALESCE(MIN(trails.duration), 0) AS min_duration FROM users LEFT JOIN trails ON users.id = trails.author OR trails.public = 1 OR EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"

		// add
		del_max_distance := &core.JSONField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "ixfa8u8m",
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
			"id": "bhn0jj40",
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
			"id": "3kujiwzh",
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
			"id": "yf3zwzn4",
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
			"id": "skvtkith",
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
			"id": "gydvo1ug",
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
		collection.Fields.RemoveById("oyvx8oaq")

		// remove
		collection.Fields.RemoveById("4xu7voeh")

		// remove
		collection.Fields.RemoveById("sqzilovv")

		// remove
		collection.Fields.RemoveById("rupidtew")

		// remove
		collection.Fields.RemoveById("2y7dbhf9")

		// remove
		collection.Fields.RemoveById("ooiwexpc")

		// remove
		collection.Fields.RemoveById("ondmfwce")

		// remove
		collection.Fields.RemoveById("kcrhzmaw")

		return app.Save(collection)
	})
}
