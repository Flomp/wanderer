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
			"id": "iz4sezoehde64wp",
			"created": "2025-02-02 12:12:48.256Z",
			"updated": "2025-02-02 12:12:48.256Z",
			"name": "integrations",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "yhqqdtrf",
					"name": "strava",
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
					"id": "bmktyzqz",
					"name": "user",
					"type": "relation",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": true,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				}
			],
			"indexes": [],
			"listRule": "@request.auth.id = user.id",
			"viewRule": "@request.auth.id = user.id",
			"createRule": "@request.auth.id = user.id",
			"updateRule": "@request.auth.id = user.id",
			"deleteRule": "@request.auth.id = user.id",
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("iz4sezoehde64wp")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
