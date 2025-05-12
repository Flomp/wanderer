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

		return app.Delete(collection)
	}, func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3208210256",
					"max": 0,
					"min": 0,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_1295301207",
					"hidden": false,
					"id": "relation1148540665",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "actor",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": true,
					"collectionId": "_pb_users_auth_",
					"hidden": false,
					"id": "_clone_kce8",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "user",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "json2215181735",
					"maxSize": 1,
					"name": "followers",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json1908379107",
					"maxSize": 1,
					"name": "following",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				}
			],
			"id": "j6w72f0kb5ivd7x",
			"indexes": [],
			"listRule": "@request.auth.id = user || (@collection.users_anonymous.id ?= user && @collection.users_anonymous.private ?= false)",
			"name": "follow_counts",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT \n  (ROW_NUMBER() OVER()) as id,\n    activitypub_actors.id as actor,\n    activitypub_actors.user as user,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM activitypub_actors\nLEFT JOIN (\n    SELECT followee AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON activitypub_actors.id = followers.actor_id\nLEFT JOIN (\n    SELECT follower AS actor_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON activitypub_actors.id = following.actor_id",
			"viewRule": "@request.auth.id = user || (@collection.users_anonymous.id ?= user && @collection.users_anonymous.private ?= false)"
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
