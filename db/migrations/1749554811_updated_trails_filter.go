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

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT activitypub_actors.id, activitypub_actors.user, COALESCE(printf(\"%.2f\", MAX(trails.distance)), 0) AS max_distance,\n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_gain)), 0) AS max_elevation_gain, \n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_loss)), 0) AS max_elevation_loss, \n  COALESCE(printf(\"%.2f\", MAX(trails.duration)), 0) AS max_duration, \n  COALESCE(printf(\"%.2f\", MIN(trails.distance)), 0) AS min_distance,   \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_gain)), 0) AS min_elevation_gain, \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_loss)), 0) AS min_elevation_loss, \n  COALESCE(printf(\"%.2f\", MIN(trails.duration)), 0) AS min_duration \nFROM activitypub_actors \n  LEFT JOIN trails ON \n  activitypub_actors.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.actor = activitypub_actors.id\n  ) GROUP BY activitypub_actors.id;"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("_clone_U0cX")

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "_clone_rnMQ",
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
		collection, err := app.FindCollectionByNameOrId("4wbv9tz5zjdrjh1")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT activitypub_actors.id, activitypub_actors.user, COALESCE(printf(\"%.2f\", MAX(trails.distance)), 0) AS max_distance,\n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_gain)), 0) AS max_elevation_gain, \n  COALESCE(printf(\"%.2f\", MAX(trails.elevation_loss)), 0) AS max_elevation_loss, \n  COALESCE(printf(\"%.2f\", MAX(trails.duration)), 0) AS max_duration, \n  COALESCE(printf(\"%.2f\", MIN(trails.distance)), 0) AS min_distance,   \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_gain)), 0) AS min_elevation_gain, \n  COALESCE(printf(\"%.2f\", MIN(trails.elevation_loss)), 0) AS min_elevation_loss, \n  COALESCE(printf(\"%.2f\", MIN(trails.duration)), 0) AS min_duration \nFROM activitypub_actors \n  LEFT JOIN trails ON \n  activitypub_actors.id = trails.author OR \n  trails.public = 1 OR \n  EXISTS (\n    SELECT 1 \n    FROM trail_share \n    WHERE trail_share.trail = trails.id \n    AND trail_share.user = activitypub_actors.user\n  ) GROUP BY activitypub_actors.id;"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "_clone_U0cX",
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
		collection.Fields.RemoveById("_clone_rnMQ")

		return app.Save(collection)
	})
}
