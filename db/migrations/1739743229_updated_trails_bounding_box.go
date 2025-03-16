package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT \n    users.id, \n    COALESCE(MAX(trails.lat), 0) AS max_lat, \n    COALESCE(MAX(trails.lon), 0) AS max_lon, \n    COALESCE(MIN(trails.lat), 0) AS min_lat, \n    COALESCE(MIN(trails.lon), 0) AS min_lon \nFROM users \nLEFT JOIN trails \n    ON users.id = trails.author \n    OR trails.public = TRUE \n    OR EXISTS (\n        SELECT 1 \n        FROM trail_share \n        WHERE trail_share.trail = trails.id \n        AND trail_share.user = users.id\n    ) \nGROUP BY users.id;"

		// remove
		collection.Fields.RemoveById("bgvxoscz")

		// remove
		collection.Fields.RemoveById("6ac8q26g")

		// remove
		collection.Fields.RemoveById("4mr8aenf")

		// remove
		collection.Fields.RemoveById("srt9ztkk")

		// add
		new_max_lat := &core.JSONField{}
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
		collection.Fields.Add(new_max_lat)

		// add
		new_max_lon := &core.JSONField{}
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
		collection.Fields.Add(new_max_lon)

		// add
		new_min_lat := &core.JSONField{}
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
		collection.Fields.Add(new_min_lat)

		// add
		new_min_lon := &core.JSONField{}
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
		collection.Fields.Add(new_min_lon)

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		collection.ViewQuery = "SELECT users.id, COALESCE(MAX(trails.lat), 0) AS max_lat, COALESCE(MAX(trails.lon), 0) AS max_lon, COALESCE(MIN(trails.lat), 0) AS min_lat, COALESCE(MIN(trails.lon), 0) AS min_lon FROM users LEFT JOIN trails ON users.id = trails.author GROUP BY users.id;"

		// add
		del_max_lat := &core.JSONField{}
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
		collection.Fields.Add(del_max_lat)

		// add
		del_max_lon := &core.JSONField{}
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
		collection.Fields.Add(del_max_lon)

		// add
		del_min_lat := &core.JSONField{}
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
		collection.Fields.Add(del_min_lat)

		// add
		del_min_lon := &core.JSONField{}
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
		collection.Fields.Add(del_min_lon)

		// remove
		collection.Fields.RemoveById("alj5aoig")

		// remove
		collection.Fields.RemoveById("mbtzxsrk")

		// remove
		collection.Fields.RemoveById("jwvtbqll")

		// remove
		collection.Fields.RemoveById("1waq3sdo")

		return app.Save(collection)
	})
}
