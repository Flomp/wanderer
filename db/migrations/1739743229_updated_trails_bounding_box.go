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

		collection, err := dao.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT \n    users.id, \n    COALESCE(MAX(trails.lat), 0) AS max_lat, \n    COALESCE(MAX(trails.lon), 0) AS max_lon, \n    COALESCE(MIN(trails.lat), 0) AS min_lat, \n    COALESCE(MIN(trails.lon), 0) AS min_lon \nFROM users \nLEFT JOIN trails \n    ON users.id = trails.author \n    OR trails.public = TRUE \n    OR EXISTS (\n        SELECT 1 \n        FROM trail_share \n        WHERE trail_share.trail = trails.id \n        AND trail_share.user = users.id\n    ) \nGROUP BY users.id;"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("bgvxoscz")

		// remove
		collection.Schema.RemoveField("6ac8q26g")

		// remove
		collection.Schema.RemoveField("4mr8aenf")

		// remove
		collection.Schema.RemoveField("srt9ztkk")

		// add
		new_max_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "alj5aoig",
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
			"id": "mbtzxsrk",
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
			"id": "jwvtbqll",
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
			"id": "1waq3sdo",
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

		options := map[string]any{}
		if err := json.Unmarshal([]byte(`{
			"query": "SELECT users.id, COALESCE(MAX(trails.lat), 0) AS max_lat, COALESCE(MAX(trails.lon), 0) AS max_lon, COALESCE(MIN(trails.lat), 0) AS min_lat, COALESCE(MIN(trails.lon), 0) AS min_lon FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"
		}`), &options); err != nil {
			return err
		}
		collection.SetOptions(options)

		// add
		del_max_lat := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "bgvxoscz",
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
			"id": "6ac8q26g",
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
			"id": "4mr8aenf",
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
			"id": "srt9ztkk",
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
		collection.Schema.RemoveField("alj5aoig")

		// remove
		collection.Schema.RemoveField("mbtzxsrk")

		// remove
		collection.Schema.RemoveField("jwvtbqll")

		// remove
		collection.Schema.RemoveField("1waq3sdo")

		return dao.SaveCollection(collection)
	})
}
