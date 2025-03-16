package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"id": "8obn1ukumze565i",
			"created": "2024-12-14 17:17:00.381Z",
			"updated": "2024-12-14 17:17:00.381Z",
			"name": "follows",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "in1traur",
					"name": "follower",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "wxwomfd5",
					"name": "followee",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				}
			],
			"indexes": [],
			"listRule": "@request.auth.id = follower.id || @request.auth.id = followee.id",
			"viewRule": "@request.auth.id = follower.id || @request.auth.id = followee.id",
			"createRule": "@request.auth.id = follower.id",
			"updateRule": "@request.auth.id = follower.id",
			"deleteRule": "@request.auth.id = follower.id",
			"options": {}
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("8obn1ukumze565i")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
