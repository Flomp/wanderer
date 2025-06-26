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

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    activitypub_actors.id, activitypub_actors.user, \n    COALESCE(MAX(trails.lat), 0) AS max_lat, \n    COALESCE(MAX(trails.lon), 0) AS max_lon, \n    COALESCE(MIN(trails.lat), 0) AS min_lat, \n    COALESCE(MIN(trails.lon), 0) AS min_lon \nFROM activitypub_actors \nLEFT JOIN trails \n    ON activitypub_actors.id = trails.author \n    OR trails.public = TRUE \n    OR EXISTS (\n        SELECT 1 \n        FROM trail_share \n        WHERE trail_share.trail = trails.id \n        AND trail_share.actor = activitypub_actors.id\n    ) \nGROUP BY activitypub_actors.id;"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("_clone_2GUq")

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "_clone_4wMO",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "user",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("urytyc428mwlbqq")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    activitypub_actors.id, activitypub_actors.user, \n    COALESCE(MAX(trails.lat), 0) AS max_lat, \n    COALESCE(MAX(trails.lon), 0) AS max_lon, \n    COALESCE(MIN(trails.lat), 0) AS min_lat, \n    COALESCE(MIN(trails.lon), 0) AS min_lon \nFROM activitypub_actors \nLEFT JOIN trails \n    ON activitypub_actors.id = trails.author \n    OR trails.public = TRUE \n    OR EXISTS (\n        SELECT 1 \n        FROM trail_share \n        WHERE trail_share.trail = trails.id \n        AND trail_share.user = activitypub_actors.id\n    ) \nGROUP BY activitypub_actors.id;"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "_clone_2GUq",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "user",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("_clone_4wMO")

		return app.Save(collection)
	})
}
