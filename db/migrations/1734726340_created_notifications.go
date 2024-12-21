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
			"id": "khrcci2uqknny8h",
			"created": "2024-12-20 17:33:49.887Z",
			"updated": "2024-12-20 20:25:40.071Z",
			"name": "notifications",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "b57prsbu",
					"name": "type",
					"type": "select",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSelect": 1,
						"values": [
							"trail_create",
							"list_create",
							"new_follower",
							"trail_comment"
						]
					}
				},
				{
					"system": false,
					"id": "1i2ycgle",
					"name": "metadata",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 2000000
					}
				},
				{
					"system": false,
					"id": "pyimxu85",
					"name": "seen",
					"type": "bool",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "tmghd4vo",
					"name": "recipient",
					"type": "relation",
					"required": true,
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
					"id": "exqo1whj",
					"name": "author",
					"type": "relation",
					"required": true,
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
			"listRule": "@request.auth.id = recipient",
			"viewRule": "@request.auth.id = recipient",
			"createRule": null,
			"updateRule": "@request.auth.id = recipient && @request.data.type = type && @request.data.metadata = metadata && @request.data.recipient = recipient && @request.data.author = author",
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("khrcci2uqknny8h")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
