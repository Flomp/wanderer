package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"id": "xku110v5a5xbufa",
			"created": "2024-05-09 17:39:57.419Z",
			"updated": "2024-05-09 17:39:57.419Z",
			"name": "users_anonymous",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "cknini6y",
					"name": "username",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				}
			],
			"indexes": [],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {
				"query": "SELECT id, username FROM users"
			}
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {

		collection, err := app.FindCollectionByNameOrId("xku110v5a5xbufa")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
