package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "j6w72f0kb5ivd7x",
			"created": "2024-12-14 18:20:18.920Z",
			"updated": "2024-12-14 18:20:18.920Z",
			"name": "follow_counts",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "w81n64il",
					"name": "followers",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				},
				{
					"system": false,
					"id": "rfnejpto",
					"name": "following",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 1
					}
				}
			],
			"indexes": [],
			"listRule": "",
			"viewRule": "",
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {
				"query": "SELECT \n    users.id,\n    COALESCE(followers.count, 0) AS followers,\n    COALESCE(following.count, 0) AS following\nFROM users\nLEFT JOIN (\n    SELECT followee AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY followee\n) AS followers ON users.id = followers.user_id\nLEFT JOIN (\n    SELECT follower AS user_id, COUNT(*) AS count\n    FROM follows\n    GROUP BY follower\n) AS following ON users.id = following.user_id;"
			}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("j6w72f0kb5ivd7x")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
