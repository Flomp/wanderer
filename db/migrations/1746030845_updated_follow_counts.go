package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("j6w72f0kb5ivd7x")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    activitypub_actors.id,\n    activitypub_actors.user as user_id,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM activitypub_actors\nLEFT JOIN (\n    SELECT followee AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON activitypub_actors.id = followers.actor_id\nLEFT JOIN (\n    SELECT follower AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON activitypub_actors.id = following.actor_id"
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(1, []byte(`{
			"cascadeDelete": true,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "_clone_mcI3",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "user_id",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("j6w72f0kb5ivd7x")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"viewQuery": "SELECT \n    activitypub_actors.id,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM activitypub_actors\nLEFT JOIN (\n    SELECT followee AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON activitypub_actors.id = followers.actor_id\nLEFT JOIN (\n    SELECT follower AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON activitypub_actors.id = following.actor_id"
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("_clone_mcI3")

		return app.Save(collection)
	})
}
