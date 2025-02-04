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
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, COALESCE(printf(\"%.2f\", MAX(trails.distance)), 0) AS max_distance,\n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_gain)), 0) AS max_elevation_gain, \n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_loss)), 0) AS max_elevation_loss, \n  COALESCE(printf(\"%.2f\", MAX(trails.duration)), 0) AS max_duration, \n  COALESCE(printf(\"%.2f\", MIN(trails.distance)), 0) AS min_distance,   \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_gain)), 0) AS min_elevation_gain, \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_loss)), 0) AS min_elevation_loss, \n  COALESCE(printf(\"%.2f\", MIN(trails.duration)), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("h18nofwq")

		// remove
		collection.Schema.RemoveField("thhrslvh")

		// remove
		collection.Schema.RemoveField("e2bbiudy")

		// remove
		collection.Schema.RemoveField("smzzpwpp")

		// remove
		collection.Schema.RemoveField("1b2horvp")

		// remove
		collection.Schema.RemoveField("m8ghv383")

		// remove
		collection.Schema.RemoveField("awhc47b7")

		// remove
		collection.Schema.RemoveField("q2houpyd")

		// add
		new_max_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(new_max_distance)

		// add
		new_max_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(new_max_elevation_gain)

		// add
		new_max_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(new_max_elevation_loss)

		// add
		new_max_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(new_max_duration)

		// add
		new_min_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(new_min_distance)

		// add
		new_min_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(new_min_elevation_gain)

		// add
		new_min_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(new_min_elevation_loss)

		// add
		new_min_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(new_min_duration)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, COALESCE(MAX(trails.distance), 0) AS max_distance,\n  COALESCE(MAX(trails.elevation_gain), 0) AS max_elevation_gain, \n  COALESCE(MAX(trails.elevation_loss), 0) AS max_elevation_loss, \n  COALESCE(MAX(trails.duration), 0) AS max_duration, \n  COALESCE(MIN(trails.distance), 0) AS min_distance,   \n  COALESCE(MIN(trails.elevation_gain), 0) AS min_elevation_gain, \n  COALESCE(MIN(trails.elevation_loss), 0) AS min_elevation_loss, \n  COALESCE(MIN(trails.duration), 0) AS min_duration \nFROM users \n  LEFT JOIN trails ON \n  users.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = users.id\n  ) GROUP BY users.id;"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// add
		del_max_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(del_max_distance)

		// add
		del_max_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(del_max_elevation_gain)

		// add
		del_max_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(del_max_elevation_loss)

		// add
		del_max_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(del_max_duration)

		// add
		del_min_distance := &schema.SchemaField{}
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
		collection.Schema.AddField(del_min_distance)

		// add
		del_min_elevation_gain := &schema.SchemaField{}
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
		collection.Schema.AddField(del_min_elevation_gain)

		// add
		del_min_elevation_loss := &schema.SchemaField{}
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
		collection.Schema.AddField(del_min_elevation_loss)

		// add
		del_min_duration := &schema.SchemaField{}
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
		collection.Schema.AddField(del_min_duration)

		// remove
		collection.Schema.RemoveField("sds7j8bk")

		// remove
		collection.Schema.RemoveField("olgpl62o")

		// remove
		collection.Schema.RemoveField("hsacrn8r")

		// remove
		collection.Schema.RemoveField("ctoyvjxp")

		// remove
		collection.Schema.RemoveField("jlwabean")

		// remove
		collection.Schema.RemoveField("w3zecui8")

		// remove
		collection.Schema.RemoveField("nby5xyd9")

		// remove
		collection.Schema.RemoveField("rn0ije3h")

		return dao.SaveCollection(collection)
	})
}
